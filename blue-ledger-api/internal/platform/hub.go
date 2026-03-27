package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// ── Message types ──────────────────────────────────────────────────────────────

const (
	MsgPresenceUpdate   = "PRESENCE_UPDATE"
	MsgChallengeInvite  = "CHALLENGE_INVITE"
	MsgChallengeAccepted = "CHALLENGE_ACCEPTED"
	MsgChallengeResult  = "CHALLENGE_RESULT"
	MsgNotification     = "NOTIFICATION"
	MsgPropsReceived    = "PROPS_RECEIVED"
	MsgMessageNew       = "MESSAGE_NEW"
	MsgPong             = "PONG"
)

// WSMessage is the envelope for all WebSocket messages.
type WSMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
}

// BroadcastMsg targets a chapter room or a specific member.
type BroadcastMsg struct {
	ChapterID string  // non-empty = broadcast to chapter room
	MemberID  string  // non-empty = targeted delivery (DM)
	Msg       WSMessage
}

// ── Client ─────────────────────────────────────────────────────────────────────

const (
	writeWait      = 10 * time.Second
	pongWait       = 90 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 4096
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins in dev; restrict in production via CORS
	},
}

// Client represents a single WebSocket connection from one member.
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	chapterID string
	memberID  string
	send      chan []byte
}

func (c *Client) readPump(rdb *redis.Client) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait)) //nolint:errcheck
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait)) //nolint:errcheck
		return nil
	})

	ctx := context.Background()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Debug().Err(err).Str("member_id", c.memberID).Msg("ws unexpected close")
			}
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "PING":
			// Refresh Redis presence TTL
			key := presenceKey(c.chapterID, c.memberID)
			rdb.Set(ctx, key, "1", pongWait) //nolint:errcheck
			// Reply with PONG
			pong, _ := json.Marshal(WSMessage{Type: MsgPong})
			select {
			case c.send <- pong:
			default:
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait)) //nolint:errcheck
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{}) //nolint:errcheck
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message) //nolint:errcheck
			// Drain queued messages in the same frame
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'}) //nolint:errcheck
				w.Write(<-c.send)    //nolint:errcheck
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait)) //nolint:errcheck
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ── Hub ────────────────────────────────────────────────────────────────────────

// Hub manages all WebSocket clients, grouped by chapter room.
type Hub struct {
	// chapter_id → set of clients
	rooms map[string]map[*Client]bool
	// member_id → client (for targeted delivery)
	byMember map[string]*Client
	mu       sync.RWMutex

	rdb        *redis.Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan BroadcastMsg
}

// NewHub creates and returns a Hub. Call hub.Run() in a goroutine.
func NewHub(rdb *redis.Client) *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		byMember:   make(map[string]*Client),
		rdb:        rdb,
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
		broadcast:  make(chan BroadcastMsg, 256),
	}
}

// Run processes hub events. Must be called in a goroutine.
func (h *Hub) Run() {
	for {
		select {

		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.chapterID] == nil {
				h.rooms[client.chapterID] = make(map[*Client]bool)
			}
			h.rooms[client.chapterID][client] = true
			h.byMember[client.memberID] = client
			h.mu.Unlock()

			// Set Redis presence key
			ctx := context.Background()
			h.rdb.Set(ctx, presenceKey(client.chapterID, client.memberID), "1", pongWait) //nolint:errcheck

			// Broadcast updated presence to chapter
			h.broadcastPresence(client.chapterID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.rooms[client.chapterID][client]; ok {
				delete(h.rooms[client.chapterID], client)
				delete(h.byMember, client.memberID)
				close(client.send)
			}
			h.mu.Unlock()

			// Remove Redis presence key
			ctx := context.Background()
			h.rdb.Del(ctx, presenceKey(client.chapterID, client.memberID)) //nolint:errcheck

			// Broadcast updated presence to chapter
			h.broadcastPresence(client.chapterID)

		case msg := <-h.broadcast:
			h.mu.RLock()
			if msg.MemberID != "" {
				// Targeted delivery
				if c, ok := h.byMember[msg.MemberID]; ok {
					h.sendToClient(c, msg.Msg)
				}
			} else if msg.ChapterID != "" {
				// Chapter-wide broadcast
				for c := range h.rooms[msg.ChapterID] {
					h.sendToClient(c, msg.Msg)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToChapter enqueues a chapter-wide broadcast.
func (h *Hub) BroadcastToChapter(chapterID string, msgType string, payload any) {
	h.broadcast <- BroadcastMsg{
		ChapterID: chapterID,
		Msg:       WSMessage{Type: msgType, Payload: payload},
	}
}

// SendToMember enqueues a targeted message to a specific member.
func (h *Hub) SendToMember(memberID string, msgType string, payload any) {
	h.broadcast <- BroadcastMsg{
		MemberID: memberID,
		Msg:      WSMessage{Type: msgType, Payload: payload},
	}
}

func (h *Hub) sendToClient(c *Client, msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
		// Client's buffer full — drop message
	}
}

// broadcastPresence reads the online member list from Redis and broadcasts it.
func (h *Hub) broadcastPresence(chapterID string) {
	ctx := context.Background()
	pattern := fmt.Sprintf("presence:%s:*", chapterID)
	keys, err := h.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return
	}

	memberIDs := make([]string, 0, len(keys))
	prefix := fmt.Sprintf("presence:%s:", chapterID)
	for _, k := range keys {
		memberIDs = append(memberIDs, k[len(prefix):])
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	msg := WSMessage{
		Type:    MsgPresenceUpdate,
		Payload: map[string]any{"online_members": memberIDs},
	}
	data, _ := json.Marshal(msg)
	for c := range h.rooms[chapterID] {
		select {
		case c.send <- data:
		default:
		}
	}
}

func presenceKey(chapterID, memberID string) string {
	return fmt.Sprintf("presence:%s:%s", chapterID, memberID)
}

// ── Echo Handlers ──────────────────────────────────────────────────────────────

// HandleWebSocket upgrades an HTTP request to WebSocket and registers the client.
func (h *Hub) HandleWebSocket(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)
	if chapterID == "" || memberID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "auth required")
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Error().Err(err).Msg("ws upgrade failed")
		return err
	}

	client := &Client{
		hub:       h,
		conn:      conn,
		chapterID: chapterID,
		memberID:  memberID,
		send:      make(chan []byte, 256),
	}

	h.register <- client

	go client.writePump()
	go client.readPump(h.rdb)

	return nil
}

// HandlePresenceList returns the current online member IDs from Redis.
func (h *Hub) HandlePresenceList(c echo.Context) error {
	chapterID := auth.GetChapterID(c)
	ctx := c.Request().Context()

	pattern := fmt.Sprintf("presence:%s:*", chapterID)
	keys, err := h.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get presence")
	}

	memberIDs := make([]string, 0, len(keys))
	prefix := fmt.Sprintf("presence:%s:", chapterID)
	for _, k := range keys {
		memberIDs = append(memberIDs, k[len(prefix):])
	}

	return c.JSON(http.StatusOK, map[string]any{"data": map[string]any{"online_members": memberIDs}})
}

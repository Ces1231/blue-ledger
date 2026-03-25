package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ces1231/blue-ledger-api/internal/auth"
	"github.com/labstack/echo/v4"
)

// Handler handles AI assistant HTTP endpoints.
type Handler struct {
	svc          Service
	anthropicKey string
	openaiKey    string
}

// NewHandler creates a new AI assistant handler.
func NewHandler(svc Service, anthropicKey, openaiKey string) *Handler {
	return &Handler{svc: svc, anthropicKey: anthropicKey, openaiKey: openaiKey}
}

// RegisterRoutes mounts AI routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMW echo.MiddlewareFunc) {
	g.Use(jwtMW)
	g.GET("/config", h.GetConfig)
	g.PUT("/config", h.UpsertConfig, auth.RoleGate("admin"))
	g.POST("/chat", h.Chat)
	g.GET("/history", h.GetHistory)
}

// GetConfig handles GET /ai/config
func (h *Handler) GetConfig(c echo.Context) error {
	cfg, err := h.svc.GetConfig(c.Request().Context(), auth.GetChapterID(c))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get ai config")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": cfg})
}

// UpsertConfig handles PUT /ai/config
func (h *Handler) UpsertConfig(c echo.Context) error {
	var req UpsertConfigInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	cfg, err := h.svc.UpsertConfig(c.Request().Context(), auth.GetChapterID(c), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update ai config")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": cfg})
}

// GetHistory handles GET /ai/history
func (h *Handler) GetHistory(c echo.Context) error {
	msgs, err := h.svc.GetHistory(c.Request().Context(), auth.GetMemberID(c))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get chat history")
	}
	return c.JSON(http.StatusOK, map[string]any{"data": msgs})
}

// Chat handles POST /ai/chat — streams SSE response
func (h *Handler) Chat(c echo.Context) error {
	var req ChatRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if strings.TrimSpace(req.Message) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "message is required")
	}

	ctx := c.Request().Context()
	chapterID := auth.GetChapterID(c)
	memberID := auth.GetMemberID(c)

	// Load config
	cfg, err := h.svc.GetConfig(ctx, chapterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load ai config")
	}
	if !cfg.Enabled {
		return echo.NewHTTPError(http.StatusForbidden, "ai assistant is not enabled for this chapter")
	}

	// Build chapter context
	chapterCtx, _ := h.svc.BuildContext(ctx, chapterID)
	systemContent := cfg.SystemPrompt
	if chapterCtx != "" {
		systemContent += "\n\nContext: " + chapterCtx
	}

	// Assemble messages
	var messages []ChatMessage
	messages = append(messages, ChatMessage{Role: "system", Content: systemContent})
	messages = append(messages, req.History...)
	messages = append(messages, ChatMessage{Role: "user", Content: req.Message})

	// Save user message
	_ = h.svc.SaveMessage(ctx, chapterID, memberID, "user", req.Message, cfg.Provider, cfg.Model)

	// Set SSE headers
	w := c.Response().Writer
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	c.Response().WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "streaming not supported")
	}

	writeSSE := func(v any) {
		b, _ := json.Marshal(v)
		fmt.Fprintf(w, "data: %s\n\n", string(b))
		flusher.Flush()
	}

	var assembled strings.Builder
	onDelta := func(text string) {
		assembled.WriteString(text)
		writeSSE(map[string]string{"delta": text})
	}

	var streamErr error
	switch cfg.Provider {
	case "openai":
		streamErr = h.svc.StreamOpenAI(ctx, h.openaiKey, cfg, messages, onDelta)
	default:
		streamErr = h.svc.StreamClaude(ctx, h.anthropicKey, cfg, messages, onDelta)
	}

	if streamErr != nil {
		writeSSE(map[string]string{"error": streamErr.Error()})
		return nil
	}

	// Save assistant response
	_ = h.svc.SaveMessage(ctx, chapterID, memberID, "assistant", assembled.String(), cfg.Provider, cfg.Model)

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
	return nil
}

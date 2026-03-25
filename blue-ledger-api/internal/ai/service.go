package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("ai config not found")
var ErrDisabled = errors.New("ai assistant is disabled for this chapter")

// AIConfig stores a chapter's AI provider configuration.
type AIConfig struct {
	ID           string    `json:"id"`
	ChapterID    string    `json:"chapter_id"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	SystemPrompt string    `json:"system_prompt"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AIMessage is a chat message stored in the DB.
type AIMessage struct {
	ID        string    `json:"id"`
	ChapterID string    `json:"chapter_id"`
	MemberID  string    `json:"member_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatMessage is used in the API request payload.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// UpsertConfigInput for updating AI config.
type UpsertConfigInput struct {
	Provider     *string `json:"provider"`
	Model        *string `json:"model"`
	SystemPrompt *string `json:"system_prompt"`
	Enabled      *bool   `json:"enabled"`
}

// ChatRequest is the incoming POST /ai/chat body.
type ChatRequest struct {
	Message string        `json:"message" validate:"required"`
	History []ChatMessage `json:"history"`
}

// Service defines AI assistant business logic.
type Service interface {
	GetConfig(ctx context.Context, chapterID string) (*AIConfig, error)
	UpsertConfig(ctx context.Context, chapterID string, input UpsertConfigInput) (*AIConfig, error)
	GetHistory(ctx context.Context, memberID string) ([]*AIMessage, error)
	SaveMessage(ctx context.Context, chapterID, memberID, role, content, provider, model string) error
	BuildContext(ctx context.Context, chapterID string) (string, error)
	StreamClaude(ctx context.Context, apiKey string, config *AIConfig, messages []ChatMessage, onDelta func(string)) error
	StreamOpenAI(ctx context.Context, apiKey string, config *AIConfig, messages []ChatMessage, onDelta func(string)) error
}

type service struct {
	pool         *pgxpool.Pool
	anthropicKey string
	openaiKey    string
}

// NewService creates a new AI assistant service.
func NewService(pool *pgxpool.Pool, anthropicKey, openaiKey string) Service {
	return &service{pool: pool, anthropicKey: anthropicKey, openaiKey: openaiKey}
}

func (s *service) GetConfig(ctx context.Context, chapterID string) (*AIConfig, error) {
	cfg := &AIConfig{}
	err := s.pool.QueryRow(ctx, `
		SELECT id, chapter_id, provider, model, system_prompt, enabled, created_at, updated_at
		FROM ai_configs WHERE chapter_id = $1
	`, chapterID).Scan(&cfg.ID, &cfg.ChapterID, &cfg.Provider, &cfg.Model, &cfg.SystemPrompt, &cfg.Enabled, &cfg.CreatedAt, &cfg.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		// Return defaults if no config exists yet.
		return &AIConfig{
			ChapterID:    chapterID,
			Provider:     "claude",
			Model:        "claude-sonnet-4-6",
			SystemPrompt: "You are a helpful assistant for a Phi Beta Sigma chapter. Be professional, supportive, and knowledgeable about fraternity operations.",
			Enabled:      true,
		}, nil
	}
	return cfg, err
}

func (s *service) UpsertConfig(ctx context.Context, chapterID string, input UpsertConfigInput) (*AIConfig, error) {
	cfg := &AIConfig{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO ai_configs (chapter_id, provider, model, system_prompt, enabled)
		VALUES ($1,
			COALESCE($2, 'claude'),
			COALESCE($3, 'claude-sonnet-4-6'),
			COALESCE($4, 'You are a helpful assistant for a Phi Beta Sigma chapter.'),
			COALESCE($5, TRUE)
		)
		ON CONFLICT (chapter_id) DO UPDATE
		SET provider     = COALESCE($2, ai_configs.provider),
		    model        = COALESCE($3, ai_configs.model),
		    system_prompt= COALESCE($4, ai_configs.system_prompt),
		    enabled      = COALESCE($5, ai_configs.enabled),
		    updated_at   = NOW()
		RETURNING id, chapter_id, provider, model, system_prompt, enabled, created_at, updated_at
	`, chapterID, input.Provider, input.Model, input.SystemPrompt, input.Enabled).
		Scan(&cfg.ID, &cfg.ChapterID, &cfg.Provider, &cfg.Model, &cfg.SystemPrompt, &cfg.Enabled, &cfg.CreatedAt, &cfg.UpdatedAt)
	return cfg, err
}

func (s *service) GetHistory(ctx context.Context, memberID string) ([]*AIMessage, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, chapter_id, member_id, role, content, provider, model, created_at
		FROM ai_messages WHERE member_id = $1 ORDER BY created_at DESC LIMIT 40
	`, memberID)
	if err != nil {
		return nil, fmt.Errorf("get ai history: %w", err)
	}
	defer rows.Close()
	var result []*AIMessage
	for rows.Next() {
		m := &AIMessage{}
		if err := rows.Scan(&m.ID, &m.ChapterID, &m.MemberID, &m.Role, &m.Content, &m.Provider, &m.Model, &m.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	// Reverse so oldest first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result, rows.Err()
}

func (s *service) SaveMessage(ctx context.Context, chapterID, memberID, role, content, provider, model string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO ai_messages (chapter_id, member_id, role, content, provider, model)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, chapterID, memberID, role, content, provider, model)
	return err
}

func (s *service) BuildContext(ctx context.Context, chapterID string) (string, error) {
	var chapterName string
	var memberCount int
	_ = s.pool.QueryRow(ctx, `SELECT name FROM chapters WHERE id = $1`, chapterID).Scan(&chapterName)
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM members WHERE chapter_id = $1 AND is_active = TRUE`, chapterID).Scan(&memberCount)

	contextStr := fmt.Sprintf("Chapter: %s | Active members: %d", chapterName, memberCount)
	return contextStr, nil
}

// ── Streaming helpers ─────────────────────────────────────────────────────────

func (s *service) StreamClaude(ctx context.Context, apiKey string, config *AIConfig, messages []ChatMessage, onDelta func(string)) error {
	// Build Anthropic messages format (no "system" role in messages array)
	type anthropicMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	var anthropicMsgs []anthropicMsg
	var systemContent string
	for _, m := range messages {
		if m.Role == "system" {
			systemContent = m.Content
			continue
		}
		anthropicMsgs = append(anthropicMsgs, anthropicMsg{Role: m.Role, Content: m.Content})
	}

	payload := map[string]any{
		"model":      config.Model,
		"max_tokens": 2048,
		"stream":     true,
		"system":     systemContent,
		"messages":   anthropicMsgs,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("claude request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("claude %d: %s", resp.StatusCode, string(errBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		if event["type"] == "content_block_delta" {
			if delta, ok := event["delta"].(map[string]any); ok {
				if text, ok := delta["text"].(string); ok {
					onDelta(text)
				}
			}
		}
	}
	return scanner.Err()
}

func (s *service) StreamOpenAI(ctx context.Context, apiKey string, config *AIConfig, messages []ChatMessage, onDelta func(string)) error {
	payload := map[string]any{
		"model":    config.Model,
		"stream":   true,
		"messages": messages,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openai %d: %s", resp.StatusCode, string(errBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk map[string]any
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		choices, ok := chunk["choices"].([]any)
		if !ok || len(choices) == 0 {
			continue
		}
		choice, ok := choices[0].(map[string]any)
		if !ok {
			continue
		}
		delta, ok := choice["delta"].(map[string]any)
		if !ok {
			continue
		}
		if text, ok := delta["content"].(string); ok {
			onDelta(text)
		}
	}
	return scanner.Err()
}

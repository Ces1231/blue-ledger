package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("chapter not found")

// ChapterSettings is the JSONB settings blob stored on the chapters table.
type ChapterSettings struct {
	ZeffyFormID         *string            `json:"zeffy_form_id,omitempty"`
	NotificationPrefs   map[string]bool    `json:"notification_prefs,omitempty"`
	XPLevels            []XPLevel          `json:"xp_levels,omitempty"`
	AllowPublicDirectory bool              `json:"allow_public_directory"`
	DarkMode            bool               `json:"dark_mode"`
	SemesterLabel       *string            `json:"semester_label,omitempty"`
}

// XPLevel defines a level threshold in the point economy.
type XPLevel struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Min   int    `json:"min"`
}

// PointEconomy holds the XP reward configuration for chapter activities.
type PointEconomy struct {
	AttendEvent    int `json:"attend_event"`
	RSVPEvent      int `json:"rsvp_event"`
	ServicePerHour int `json:"service_per_hour"`
	Props          int `json:"props"`
	DuesPaid       int `json:"dues_paid"`
	BadgeBonus     int `json:"badge_bonus"`
}

// UpdateSettingsInput holds data for updating chapter settings.
type UpdateSettingsInput struct {
	AllowPublicDirectory *bool           `json:"allow_public_directory"`
	DarkMode             *bool           `json:"dark_mode"`
	NotificationPrefs    map[string]bool `json:"notification_prefs"`
	SemesterLabel        *string         `json:"semester_label"`
}

// ZeffyConfigInput holds Zeffy integration settings.
type ZeffyConfigInput struct {
	FormID string `json:"form_id" validate:"required"`
}

// Service defines the settings business logic interface.
type Service interface {
	GetSettings(ctx context.Context, chapterID string) (*ChapterSettings, error)
	UpdateSettings(ctx context.Context, chapterID string, input UpdateSettingsInput) (*ChapterSettings, error)
	GetPointEconomy(ctx context.Context, chapterID string) (*PointEconomy, error)
	UpdatePointEconomy(ctx context.Context, chapterID string, input PointEconomy) (*PointEconomy, error)
	SaveZeffyConfig(ctx context.Context, chapterID string, input ZeffyConfigInput) (*ChapterSettings, error)
}

type service struct {
	pool *pgxpool.Pool
}

// NewService creates a new settings service.
func NewService(pool *pgxpool.Pool) Service {
	return &service{pool: pool}
}

func (s *service) GetSettings(ctx context.Context, chapterID string) (*ChapterSettings, error) {
	var raw json.RawMessage
	err := s.pool.QueryRow(ctx,
		`SELECT settings FROM chapters WHERE id = $1 AND deleted_at IS NULL`, chapterID,
	).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get chapter settings: %w", err)
	}

	var cfg ChapterSettings
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return &ChapterSettings{}, nil // return defaults if JSONB is malformed
	}
	return &cfg, nil
}

func (s *service) UpdateSettings(ctx context.Context, chapterID string, input UpdateSettingsInput) (*ChapterSettings, error) {
	current, err := s.GetSettings(ctx, chapterID)
	if err != nil {
		return nil, err
	}

	if input.AllowPublicDirectory != nil {
		current.AllowPublicDirectory = *input.AllowPublicDirectory
	}
	if input.DarkMode != nil {
		current.DarkMode = *input.DarkMode
	}
	if input.NotificationPrefs != nil {
		current.NotificationPrefs = input.NotificationPrefs
	}
	if input.SemesterLabel != nil {
		current.SemesterLabel = input.SemesterLabel
	}

	return s.saveSettings(ctx, chapterID, current)
}

func (s *service) GetPointEconomy(ctx context.Context, chapterID string) (*PointEconomy, error) {
	cfg, err := s.GetSettings(ctx, chapterID)
	if err != nil {
		return nil, err
	}

	// Return defaults if not explicitly configured
	pe := defaultPointEconomy()

	// XP levels are stored in settings.xp_levels; point economy is a separate key
	// For now, store/retrieve from the settings JSONB via a dedicated key
	var raw json.RawMessage
	_ = s.pool.QueryRow(ctx,
		`SELECT settings->'point_economy' FROM chapters WHERE id = $1`, chapterID,
	).Scan(&raw)

	if raw != nil {
		_ = json.Unmarshal(raw, pe)
	}
	_ = cfg
	return pe, nil
}

func (s *service) UpdatePointEconomy(ctx context.Context, chapterID string, input PointEconomy) (*PointEconomy, error) {
	peJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshal point economy: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE chapters
		SET settings = jsonb_set(settings, '{point_economy}', $2::jsonb, true),
			updated_at = NOW()
		WHERE id = $1
	`, chapterID, string(peJSON))
	if err != nil {
		return nil, fmt.Errorf("update point economy: %w", err)
	}
	return &input, nil
}

func (s *service) SaveZeffyConfig(ctx context.Context, chapterID string, input ZeffyConfigInput) (*ChapterSettings, error) {
	current, err := s.GetSettings(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	current.ZeffyFormID = &input.FormID
	return s.saveSettings(ctx, chapterID, current)
}

func (s *service) saveSettings(ctx context.Context, chapterID string, cfg *ChapterSettings) (*ChapterSettings, error) {
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal settings: %w", err)
	}

	_, err = s.pool.Exec(ctx,
		`UPDATE chapters SET settings = $2, updated_at = NOW() WHERE id = $1`,
		chapterID, string(raw),
	)
	if err != nil {
		return nil, fmt.Errorf("save settings: %w", err)
	}
	return cfg, nil
}

func defaultPointEconomy() *PointEconomy {
	return &PointEconomy{
		AttendEvent:    25,
		RSVPEvent:      5,
		ServicePerHour: 10,
		Props:          10,
		DuesPaid:       50,
		BadgeBonus:     0,
	}
}

package members

import (
	"context"
	"fmt"
)

// Service defines the members domain business logic interface.
type Service interface {
	List(ctx context.Context, chapterID string, isSysadmin bool, page, perPage int) ([]*Member, int, error)
	Get(ctx context.Context, chapterID string, isSysadmin bool, memberID string) (*Member, error)
	Create(ctx context.Context, chapterID string, input CreateInput) (*Member, error)
	Update(ctx context.Context, chapterID string, isSysadmin bool, memberID string, input UpdateInput, requestorRole string) (*Member, error)
	Delete(ctx context.Context, chapterID string, isSysadmin bool, memberID string) error
	Reactivate(ctx context.Context, chapterID string, isSysadmin bool, memberID string) (*Member, error)
	GetXPHistory(ctx context.Context, chapterID string, isSysadmin bool, memberID string, page, perPage int) ([]*XPHistoryEntry, int, error)
}

type service struct {
	repo *Repository
}

// NewService creates a new members service.
func NewService(repo *Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, chapterID string, isSysadmin bool, page, perPage int) ([]*Member, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	return s.repo.List(ctx, chapterID, isSysadmin, page, perPage)
}

func (s *service) Get(ctx context.Context, chapterID string, isSysadmin bool, memberID string) (*Member, error) {
	m, err := s.repo.GetByID(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("get member: %w", err)
	}
	if m == nil {
		return nil, nil
	}
	// Verify the member belongs to the requesting chapter (unless sysadmin)
	// Defense in depth — RLS also enforces this
	if !isSysadmin && m.ChapterID != chapterID {
		return nil, nil
	}
	return m, nil
}

func (s *service) Create(ctx context.Context, chapterID string, input CreateInput) (*Member, error) {
	m, err := s.repo.Create(ctx, chapterID, input)
	if err != nil {
		return nil, fmt.Errorf("create member: %w", err)
	}
	return m, nil
}

func (s *service) Update(ctx context.Context, chapterID string, isSysadmin bool, memberID string, input UpdateInput, requestorRole string) (*Member, error) {
	// Verify member belongs to chapter (unless sysadmin)
	m, err := s.repo.GetByID(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("verify member: %w", err)
	}
	if m == nil {
		return nil, nil
	}
	if !isSysadmin && m.ChapterID != chapterID {
		return nil, fmt.Errorf("member not in your chapter")
	}

	// Non-admins can only update their own profile fields; role/status/dues_status are admin-only
	if requestorRole != "admin" && requestorRole != "sysadmin" {
		input.Role = nil
		input.Status = nil
		input.DuesStatus = nil
	}

	m, err = s.repo.Update(ctx, memberID, input)
	if err != nil {
		return nil, fmt.Errorf("update member: %w", err)
	}
	return m, nil
}

func (s *service) Delete(ctx context.Context, chapterID string, isSysadmin bool, memberID string) error {
	// Verify member belongs to chapter (unless sysadmin)
	m, err := s.repo.GetByID(ctx, memberID)
	if err != nil {
		return fmt.Errorf("verify member: %w", err)
	}
	if m == nil {
		return fmt.Errorf("member not found")
	}
	if !isSysadmin && m.ChapterID != chapterID {
		return fmt.Errorf("member not in your chapter")
	}
	return s.repo.SoftDelete(ctx, memberID)
}

func (s *service) Reactivate(ctx context.Context, chapterID string, isSysadmin bool, memberID string) (*Member, error) {
	// Verify member belongs to chapter (unless sysadmin)
	// Note: Reactivate uses a special query that doesn't filter by deleted_at
	m, err := s.repo.GetByIDIncludeDeleted(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("verify member: %w", err)
	}
	if m == nil {
		return nil, nil
	}
	if !isSysadmin && m.ChapterID != chapterID {
		return nil, fmt.Errorf("member not in your chapter")
	}
	return s.repo.Reactivate(ctx, memberID)
}

func (s *service) GetXPHistory(ctx context.Context, chapterID string, isSysadmin bool, memberID string, page, perPage int) ([]*XPHistoryEntry, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	// Verify member belongs to chapter (unless sysadmin)
	m, err := s.repo.GetByID(ctx, memberID)
	if err != nil {
		return nil, 0, fmt.Errorf("verify member: %w", err)
	}
	if m == nil {
		return nil, 0, fmt.Errorf("member not found")
	}
	if !isSysadmin && m.ChapterID != chapterID {
		return nil, 0, fmt.Errorf("member not in your chapter")
	}
	return s.repo.GetXPHistory(ctx, memberID, page, perPage)
}

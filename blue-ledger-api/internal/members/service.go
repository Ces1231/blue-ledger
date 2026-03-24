package members

import (
	"context"
	"fmt"
)

// Service defines the members domain business logic interface.
type Service interface {
	List(ctx context.Context, chapterID string, page, perPage int) ([]*Member, int, error)
	Get(ctx context.Context, chapterID, memberID string) (*Member, error)
	Update(ctx context.Context, chapterID, memberID string, input UpdateInput, requestorRole string) (*Member, error)
	Delete(ctx context.Context, chapterID, memberID string) error
	GetXPHistory(ctx context.Context, chapterID, memberID string, page, perPage int) ([]*XPHistoryEntry, int, error)
}

type service struct {
	repo *Repository
}

// NewService creates a new members service.
func NewService(repo *Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, chapterID string, page, perPage int) ([]*Member, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	return s.repo.List(ctx, chapterID, page, perPage)
}

func (s *service) Get(ctx context.Context, chapterID, memberID string) (*Member, error) {
	m, err := s.repo.GetByID(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("get member: %w", err)
	}
	if m == nil {
		return nil, nil
	}
	// Verify the member belongs to the requesting chapter (defense in depth — RLS also enforces this)
	if m.ChapterID != chapterID {
		return nil, nil
	}
	return m, nil
}

func (s *service) Update(ctx context.Context, chapterID, memberID string, input UpdateInput, requestorRole string) (*Member, error) {
	// Non-admins can only update their own profile fields; role/status/dues_status are admin-only
	if requestorRole != "admin" && requestorRole != "sysadmin" {
		input.Role = nil
		input.Status = nil
		input.DuesStatus = nil
	}

	m, err := s.repo.Update(ctx, memberID, input)
	if err != nil {
		return nil, fmt.Errorf("update member: %w", err)
	}
	return m, nil
}

func (s *service) Delete(ctx context.Context, chapterID, memberID string) error {
	return s.repo.SoftDelete(ctx, memberID)
}

func (s *service) GetXPHistory(ctx context.Context, chapterID, memberID string, page, perPage int) ([]*XPHistoryEntry, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	return s.repo.GetXPHistory(ctx, memberID, page, perPage)
}

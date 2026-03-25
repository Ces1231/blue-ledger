package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// mockHealthService implements Service for testing.
type mockHealthService struct {
	overview   *ChapterHealth
	attendance []*AttendanceStat
	xpStats    []*XPStat
	svcStats   []*ServiceStat
	duesStats  []*DuesStat
}

func (m *mockHealthService) GetOverview(ctx context.Context, chapterID string) (*ChapterHealth, error) {
	if m.overview != nil {
		return m.overview, nil
	}
	return &ChapterHealth{ActiveMembers: 10, TotalMembers: 12, AttendanceRate: 0.83}, nil
}

func (m *mockHealthService) GetAttendance(ctx context.Context, chapterID string, limit int) ([]*AttendanceStat, error) {
	return m.attendance, nil
}

func (m *mockHealthService) GetXPDistribution(ctx context.Context, chapterID string) ([]*XPStat, error) {
	return m.xpStats, nil
}

func (m *mockHealthService) GetServiceStats(ctx context.Context, chapterID string, months int) ([]*ServiceStat, error) {
	return m.svcStats, nil
}

func (m *mockHealthService) GetDuesBreakdown(ctx context.Context, chapterID string) ([]*DuesStat, error) {
	return m.duesStats, nil
}

func setupHealthEcho(method, path string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("chapter_id", "chapter-test")
	c.Set("role", "admin")
	return c, rec
}

func TestHealthHandler_GetOverview(t *testing.T) {
	h := NewHandler(&mockHealthService{
		overview: &ChapterHealth{
			ActiveMembers:  15,
			TotalMembers:   18,
			AttendanceRate: 0.75,
			AvgXP:          320.5,
		},
	})

	c, rec := setupHealthEcho(http.MethodGet, "/health")
	if err := h.GetOverview(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d want 200", rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := resp["data"]; !ok {
		t.Error("response missing 'data' key")
	}
}

func TestHealthHandler_GetAttendance_Empty(t *testing.T) {
	h := NewHandler(&mockHealthService{
		attendance: []*AttendanceStat{},
	})

	c, rec := setupHealthEcho(http.MethodGet, "/health/attendance")
	if err := h.GetAttendance(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d want 200", rec.Code)
	}
}

func TestHealthHandler_GetXPDistribution(t *testing.T) {
	h := NewHandler(&mockHealthService{
		xpStats: []*XPStat{
			{Level: "Neophyte", Count: 5, AvgXP: 250},
			{Level: "Bronze Varsity", Count: 3, AvgXP: 720},
		},
	})

	c, rec := setupHealthEcho(http.MethodGet, "/health/xp")
	if err := h.GetXPDistribution(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d want 200", rec.Code)
	}
}

func TestHealthHandler_GetDuesBreakdown(t *testing.T) {
	h := NewHandler(&mockHealthService{
		duesStats: []*DuesStat{},
	})

	c, rec := setupHealthEcho(http.MethodGet, "/health/dues")
	if err := h.GetDuesBreakdown(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d want 200", rec.Code)
	}
}

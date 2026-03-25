package health

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ChapterHealth aggregates all health metrics for a chapter.
type ChapterHealth struct {
	AttendanceRate float64 `json:"attendance_rate"`
	AvgXP          float64 `json:"avg_xp"`
	TotalServiceHours float64 `json:"total_service_hours"`
	DuesPaidRate   float64 `json:"dues_paid_rate"`
	ActiveMembers  int     `json:"active_members"`
	TotalMembers   int     `json:"total_members"`
}

// AttendanceStat is per-event attendance data.
type AttendanceStat struct {
	EventName    string  `json:"event_name"`
	EventDate    string  `json:"event_date"`
	TotalRSVPs   int     `json:"total_rsvps"`
	TotalAttended int    `json:"total_attended"`
	Rate         float64 `json:"rate"`
}

// XPStat is XP distribution data.
type XPStat struct {
	Level    string `json:"level"`
	Count    int    `json:"count"`
	AvgXP    float64 `json:"avg_xp"`
}

// ServiceStat is service hours by month.
type ServiceStat struct {
	Period string  `json:"period"`
	Hours  float64 `json:"hours"`
	Count  int     `json:"count"`
}

// DuesStat is dues payment breakdown.
type DuesStat struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// Service defines the health analytics interface.
type Service interface {
	GetOverview(ctx context.Context, chapterID string) (*ChapterHealth, error)
	GetAttendance(ctx context.Context, chapterID string, limit int) ([]*AttendanceStat, error)
	GetXPDistribution(ctx context.Context, chapterID string) ([]*XPStat, error)
	GetServiceStats(ctx context.Context, chapterID string, months int) ([]*ServiceStat, error)
	GetDuesBreakdown(ctx context.Context, chapterID string) ([]*DuesStat, error)
}

type service struct {
	pool *pgxpool.Pool
}

// NewService creates a new health analytics service.
func NewService(pool *pgxpool.Pool) Service {
	return &service{pool: pool}
}

func (s *service) GetOverview(ctx context.Context, chapterID string) (*ChapterHealth, error) {
	h := &ChapterHealth{}
	err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'active') AS active_members,
			COUNT(*) AS total_members
		FROM members
		WHERE chapter_id = $1
	`, chapterID).Scan(&h.ActiveMembers, &h.TotalMembers)
	if err != nil {
		return nil, fmt.Errorf("get overview member counts: %w", err)
	}

	// Average XP
	_ = s.pool.QueryRow(ctx,
		`SELECT COALESCE(AVG(xp_total), 0) FROM members WHERE chapter_id = $1 AND status = 'active'`,
		chapterID,
	).Scan(&h.AvgXP)

	// Attendance rate (last 90 days)
	var attended, rsvpd int
	_ = s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE r.attended = TRUE),
			COUNT(*)
		FROM rsvps r
		JOIN events e ON e.id = r.event_id
		WHERE e.chapter_id = $1 AND e.event_date >= NOW() - INTERVAL '90 days'
	`, chapterID).Scan(&attended, &rsvpd)
	if rsvpd > 0 {
		h.AttendanceRate = float64(attended) / float64(rsvpd) * 100
	}

	// Total service hours (current semester)
	_ = s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(hours), 0) FROM service_log WHERE chapter_id = $1 AND verified = TRUE`,
		chapterID,
	).Scan(&h.TotalServiceHours)

	// Dues paid rate
	var paidCount, totalDues int
	_ = s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'paid'),
			COUNT(*)
		FROM dues_records
		WHERE chapter_id = $1
	`, chapterID).Scan(&paidCount, &totalDues)
	if totalDues > 0 {
		h.DuesPaidRate = float64(paidCount) / float64(totalDues) * 100
	}

	return h, nil
}

func (s *service) GetAttendance(ctx context.Context, chapterID string, limit int) ([]*AttendanceStat, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}
	rows, err := s.pool.Query(ctx, `
		SELECT
			e.name,
			e.event_date::text,
			COUNT(r.id) FILTER (WHERE r.status = 'yes') AS total_rsvps,
			COUNT(r.id) FILTER (WHERE r.attended = TRUE) AS total_attended,
			CASE WHEN COUNT(r.id) FILTER (WHERE r.status = 'yes') > 0
				THEN (COUNT(r.id) FILTER (WHERE r.attended = TRUE))::float /
					(COUNT(r.id) FILTER (WHERE r.status = 'yes'))::float * 100
				ELSE 0
			END AS rate
		FROM events e
		LEFT JOIN rsvps r ON r.event_id = e.id
		WHERE e.chapter_id = $1 AND e.event_date <= NOW()
		GROUP BY e.id
		ORDER BY e.event_date DESC
		LIMIT $2
	`, chapterID, limit)
	if err != nil {
		return nil, fmt.Errorf("get attendance stats: %w", err)
	}
	defer rows.Close()

	var result []*AttendanceStat
	for rows.Next() {
		stat := &AttendanceStat{}
		if err := rows.Scan(&stat.EventName, &stat.EventDate, &stat.TotalRSVPs, &stat.TotalAttended, &stat.Rate); err != nil {
			return nil, fmt.Errorf("scan attendance stat: %w", err)
		}
		result = append(result, stat)
	}
	return result, rows.Err()
}

func (s *service) GetXPDistribution(ctx context.Context, chapterID string) ([]*XPStat, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT level, COUNT(*) AS count, COALESCE(AVG(xp_total), 0) AS avg_xp
		FROM members
		WHERE chapter_id = $1 AND status = 'active'
		GROUP BY level
		ORDER BY avg_xp DESC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("get XP distribution: %w", err)
	}
	defer rows.Close()

	var result []*XPStat
	for rows.Next() {
		stat := &XPStat{}
		if err := rows.Scan(&stat.Level, &stat.Count, &stat.AvgXP); err != nil {
			return nil, fmt.Errorf("scan XP stat: %w", err)
		}
		result = append(result, stat)
	}
	return result, rows.Err()
}

func (s *service) GetServiceStats(ctx context.Context, chapterID string, months int) ([]*ServiceStat, error) {
	if months < 1 || months > 24 {
		months = 6
	}
	rows, err := s.pool.Query(ctx, `
		SELECT
			TO_CHAR(DATE_TRUNC('month', service_date::date), 'YYYY-MM') AS period,
			COALESCE(SUM(hours), 0) AS hours,
			COUNT(*) AS count
		FROM service_log
		WHERE chapter_id = $1
			AND verified = TRUE
			AND service_date >= (NOW() - ($2::int * INTERVAL '1 month'))::text::date
		GROUP BY DATE_TRUNC('month', service_date::date)
		ORDER BY period DESC
	`, chapterID, months)
	if err != nil {
		return nil, fmt.Errorf("get service stats: %w", err)
	}
	defer rows.Close()

	var result []*ServiceStat
	for rows.Next() {
		stat := &ServiceStat{}
		if err := rows.Scan(&stat.Period, &stat.Hours, &stat.Count); err != nil {
			return nil, fmt.Errorf("scan service stat: %w", err)
		}
		result = append(result, stat)
	}
	return result, rows.Err()
}

func (s *service) GetDuesBreakdown(ctx context.Context, chapterID string) ([]*DuesStat, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT status, COUNT(*) AS count
		FROM dues_records
		WHERE chapter_id = $1
		GROUP BY status
		ORDER BY count DESC
	`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("get dues breakdown: %w", err)
	}
	defer rows.Close()

	var result []*DuesStat
	for rows.Next() {
		stat := &DuesStat{}
		if err := rows.Scan(&stat.Status, &stat.Count); err != nil {
			return nil, fmt.Errorf("scan dues stat: %w", err)
		}
		result = append(result, stat)
	}
	return result, rows.Err()
}

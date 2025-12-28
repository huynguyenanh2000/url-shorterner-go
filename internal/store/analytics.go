package store

import (
	"context"
	"database/sql"
	"time"
)

type URLAnalytics struct {
	ID        uint64    `json:"id"`
	URLID     uint64    `json:"url_id"`
	ClickTime time.Time `json:"click_time"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Referrer  string    `json:"referrer"`
}

type StatsSummary struct {
	TotalClicks int            `json:"total_clicks"`
	Clicks24h   int            `json:"clicks_24h"`
	TopReferrer string         `json:"top_referrer"`
	DeviceStats map[string]int `json:"device_stats"`
}

type AnalyticsStore struct {
	db *sql.DB
}

func (s *AnalyticsStore) Create(ctx context.Context, analytics *URLAnalytics) error {
	query := `
		INSERT INTO url_analytics (url_id, ip_address, user_agent, referrer)
		VALUES (?, ?, ?, ?)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		analytics.URLID,
		analytics.IPAddress,
		analytics.UserAgent,
		analytics.Referrer,
	)
	return err
}

func (s *AnalyticsStore) GetStatsByURLID(ctx context.Context, urlID uint64) (*StatsSummary, error) {
	// Initialize the summary struct with an empty map for DeviceStats
	summary := &StatsSummary{
		DeviceStats: make(map[string]int),
	}

	// 1. Fetch Aggregated Counts and Top Referrer
	// This query utilizes the composite index (url_id, click_time) for maximum speed.
	// We use a subquery for top_referrer to keep the main count logic clean.
	summaryQuery := `
			SELECT
				COUNT(id) AS total_clicks,
				COUNT(CASE WHEN click_time >= NOW() - INTERVAL 1 DAY THEN 1 END) AS clicks_24h,
				COALESCE(
					(SELECT referrer FROM url_analytics
					 WHERE url_id = ? AND referrer IS NOT NULL AND referrer <> ''
					 GROUP BY referrer ORDER BY COUNT(*) DESC LIMIT 1),
					'Direct'
				) AS top_referrer
			FROM url_analytics
			WHERE url_id = ?
		`

	// Use a timeout to prevent long-running queries from blocking the application
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, summaryQuery, urlID, urlID).Scan(
		&summary.TotalClicks,
		&summary.Clicks24h,
		&summary.TopReferrer,
	)
	if err != nil {
		// If no analytics records exist for this ID, SQL might return an error or 0s.
		// Since we use urlID from a validated context, we treat no rows as 0 stats.
		if err == sql.ErrNoRows {
			return summary, nil
		}
		return nil, err
	}

	// 2. Fetch Device Distribution
	// We categorize the User-Agent strings into 'Mobile' or 'Desktop'.
	deviceQuery := `
			SELECT
				CASE
					WHEN user_agent LIKE '%Mobi%' THEN 'Mobile'
					ELSE 'Desktop'
				END AS device_type,
				COUNT(*)
			FROM url_analytics
			WHERE url_id = ?
			GROUP BY device_type
		`

	rows, err := s.db.QueryContext(ctx, deviceQuery, urlID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var deviceType string
		var count int
		if err := rows.Scan(&deviceType, &count); err != nil {
			return nil, err
		}
		summary.DeviceStats[deviceType] = count
	}

	// Check for any errors that occurred during the iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return summary, nil
}

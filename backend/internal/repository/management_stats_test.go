package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestProxyRepositoryGetStats(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := newProxyRepositoryWithSQL(nil, db)

	mock.ExpectQuery("WITH proxy_accounts AS").
		WithArgs(int64(9), service.StatusActive).
		WillReturnRows(sqlmock.NewRows([]string{
			"total_accounts", "active_accounts", "total_requests", "success_rate", "average_latency",
		}).AddRow(int64(4), int64(3), int64(25), 80.0, 125.5))

	stats, err := repo.GetStats(context.Background(), 9)
	require.NoError(t, err)
	require.Equal(t, int64(4), stats.TotalAccounts)
	require.Equal(t, int64(3), stats.ActiveAccounts)
	require.Equal(t, int64(25), stats.TotalRequests)
	require.Equal(t, 80.0, stats.SuccessRate)
	require.Equal(t, 125.5, stats.AverageLatency)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageLogRepositoryGetRealtimeMetrics(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &usageLogRepository{sql: db}
	start := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	end := start.Add(time.Minute)

	mock.ExpectQuery("WITH success_stats AS").
		WithArgs(start, end).
		WillReturnRows(sqlmock.NewRows([]string{
			"requests_per_minute", "average_response_time", "error_rate",
		}).AddRow(int64(120), 230.25, 2.5))

	metrics, err := repo.GetRealtimeMetrics(context.Background(), start, end)
	require.NoError(t, err)
	require.Equal(t, int64(120), metrics.RequestsPerMinute)
	require.Equal(t, 230.25, metrics.AverageResponseTime)
	require.Equal(t, 2.5, metrics.ErrorRate)
	require.NoError(t, mock.ExpectationsWereMet())
}

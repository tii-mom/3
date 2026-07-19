package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type managementStatsUsageRepo struct {
	service.UsageLogRepository
}

func (r *managementStatsUsageRepo) GetGroupStatsWithFilters(
	_ context.Context,
	start, end time.Time,
	_, _, _, groupID int64,
	_ *int16,
	_ *bool,
	_ *int8,
) ([]usagestats.GroupStat, error) {
	return []usagestats.GroupStat{{
		GroupID:  groupID,
		Requests: 42,
		Cost:     12.75,
	}}, nil
}

func (r *managementStatsUsageRepo) GetRealtimeMetrics(
	_ context.Context,
	start, end time.Time,
) (*usagestats.RealtimeMetrics, error) {
	return &usagestats.RealtimeMetrics{
		RequestsPerMinute:   90,
		AverageResponseTime: 155.5,
		ErrorRate:           4.25,
	}, nil
}

func TestGroupHandlerGetStatsCombinesKeysAndUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	adminSvc := newStubAdminService()
	dashboardSvc := service.NewDashboardService(&managementStatsUsageRepo{}, nil, nil, nil)
	handler := NewGroupHandler(adminSvc, dashboardSvc, nil)
	router := gin.New()
	router.GET("/groups/:id/stats", handler.GetStats)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/groups/2/stats", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{
		"code": 0,
		"message": "success",
		"data": {
			"total_api_keys": 1,
			"active_api_keys": 1,
			"total_requests": 42,
			"total_cost": 12.75
		}
	}`, rec.Body.String())
}

func TestDashboardHandlerGetRealtimeMetricsUsesRepositoryValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dashboardSvc := service.NewDashboardService(&managementStatsUsageRepo{}, nil, nil, nil)
	handler := NewDashboardHandler(dashboardSvc, nil)
	router := gin.New()
	router.GET("/dashboard/realtime", handler.GetRealtimeMetrics)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dashboard/realtime", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"requests_per_minute":90`)
	require.Contains(t, rec.Body.String(), `"average_response_time":155.5`)
	require.Contains(t, rec.Body.String(), `"error_rate":4.25`)
}

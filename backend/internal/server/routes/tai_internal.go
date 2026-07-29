package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterTAIInternalRoutes registers minimal server-to-server endpoints
// for TAI Protocol Phase 0 (platform pool account model).
//
// Phase 0: TAI backend proxies all pet API calls through ONE platform key.
// These endpoints are for reconciliation and monitoring only.
func RegisterTAIInternalRoutes(
	v1 *gin.RouterGroup,
	taiHandler *handler.TAIInternalHandler,
	internalSecret string,
) {
	tai := v1.Group("/internal/tai")
	tai.Use(middleware.InternalSecretGuard(internalSecret))
	{
		// Report aggregated usage from TAI backend (for reconciliation)
		tai.POST("/usage-report", taiHandler.UsageReport)

		// Get platform account balance and status
		tai.GET("/account-status", taiHandler.AccountStatus)

		// Alert: platform balance running low
		tai.GET("/balance-check", taiHandler.BalanceCheck)
	}
}

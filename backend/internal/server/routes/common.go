package routes

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RegisterCommonRoutes 注册通用路由（健康检查、状态等）
func RegisterCommonRoutes(r *gin.Engine, db *sql.DB, redisClient *redis.Client) {
	// Liveness intentionally has no dependency checks.
	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	readiness := readinessHandler(db, redisClient)
	r.GET("/health", readiness)
	r.GET("/health/ready", readiness)

	// Claude Code 遥测日志（忽略，直接返回200）
	r.POST("/api/event_logging/batch", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Setup status endpoint (always returns needs_setup: false in normal mode)
	// This is used by the frontend to detect when the service has restarted after setup
	r.GET("/setup/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"needs_setup": false,
				"step":        "completed",
			},
		})
	})
}

func readinessHandler(db *sql.DB, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		components := gin.H{}
		ready := true
		if db == nil || db.PingContext(ctx) != nil {
			components["postgres"] = "unavailable"
			ready = false
		} else {
			components["postgres"] = "ok"
		}
		if redisClient == nil || redisClient.Ping(ctx).Err() != nil {
			components["redis"] = "unavailable"
			ready = false
		} else {
			components["redis"] = "ok"
		}

		statusCode := http.StatusOK
		status := "ok"
		if !ready {
			statusCode = http.StatusServiceUnavailable
			status = "unavailable"
		}
		c.JSON(statusCode, gin.H{"status": status, "components": components})
	}
}

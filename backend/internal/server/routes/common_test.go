package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestHealthRoutesDistinguishLivenessAndReadiness(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	mock.ExpectPing()

	mini := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	defer func() { _ = redisClient.Close() }()

	router := gin.New()
	RegisterCommonRoutes(router, db, redisClient)

	readyRecorder := httptest.NewRecorder()
	router.ServeHTTP(readyRecorder, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	require.Equal(t, http.StatusOK, readyRecorder.Code)
	require.JSONEq(t, `{"status":"ok","components":{"postgres":"ok","redis":"ok"}}`, readyRecorder.Body.String())
	require.NoError(t, mock.ExpectationsWereMet())

	require.NoError(t, redisClient.Close())
	failedRecorder := httptest.NewRecorder()
	router.ServeHTTP(failedRecorder, httptest.NewRequest(http.MethodGet, "/health", nil))
	require.Equal(t, http.StatusServiceUnavailable, failedRecorder.Code)

	liveRecorder := httptest.NewRecorder()
	router.ServeHTTP(liveRecorder, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	require.Equal(t, http.StatusOK, liveRecorder.Code)

}

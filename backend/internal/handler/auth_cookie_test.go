package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestBrowserAuthCookiesAndCSRFValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AuthHandler{cfg: &config.Config{}}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)

	csrfToken, err := h.setBrowserAuthCookies(c, "refresh-token")
	require.NoError(t, err)
	require.NotEmpty(t, csrfToken)

	cookies := recorder.Result().Cookies()
	require.Len(t, cookies, 2)
	for _, cookie := range cookies {
		require.Equal(t, authCookiePath, cookie.Path)
		require.True(t, cookie.HttpOnly)
		require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
		require.Empty(t, cookie.Domain)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	request.Header.Set(csrfTokenHeader, csrfToken)
	c.Request = request
	require.NoError(t, validateBrowserCSRF(c))

	c.Request.Header.Set(csrfTokenHeader, "wrong-token")
	require.Error(t, validateBrowserCSRF(c))
}

func TestOAuthTokenPairResponseUsesCookieTransportWithoutJSONRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AuthHandler{cfg: &config.Config{}}
	tokenPair := &service.TokenPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    3600,
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/oauth/pending/exchange", nil)
	c.Request.Header.Set(authTransportHeader, authTransportCookie)

	h.writeOAuthTokenPairResponse(c, tokenPair)

	require.Equal(t, http.StatusOK, recorder.Code)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	require.Equal(t, "access-token", payload["access_token"])
	require.NotEmpty(t, payload["csrf_token"])
	require.NotContains(t, payload, "refresh_token")
	require.Len(t, recorder.Result().Cookies(), 2)
}

func TestOAuthTokenPairResponsePreservesLegacyJSONTransport(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AuthHandler{cfg: &config.Config{}}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/oauth/pending/exchange", nil)

	h.writeOAuthTokenPairResponse(c, &service.TokenPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    3600,
	})

	var payload map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	require.Equal(t, "refresh-token", payload["refresh_token"])
	require.Empty(t, recorder.Result().Cookies())
}

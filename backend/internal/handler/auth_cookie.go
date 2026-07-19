package handler

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/gin-gonic/gin"
)

const (
	authTransportHeader = "X-Auth-Transport"
	authTransportCookie = "cookie"
	refreshTokenCookie  = "sub2api_refresh"
	csrfTokenCookie     = "sub2api_csrf"
	csrfTokenHeader     = "X-CSRF-Token"
	authCookiePath      = "/api/v1/auth"
)

func cookieAuthRequested(c *gin.Context) bool {
	return c != nil && strings.EqualFold(strings.TrimSpace(c.GetHeader(authTransportHeader)), authTransportCookie)
}

func (h *AuthHandler) authCookieSecure(c *gin.Context) bool {
	if h != nil && h.cfg != nil && strings.EqualFold(strings.TrimSpace(h.cfg.Server.Mode), "release") {
		return true
	}
	return c != nil && (c.Request.TLS != nil || strings.EqualFold(strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")), "https"))
}

func (h *AuthHandler) setBrowserAuthCookies(c *gin.Context, refreshToken string) (string, error) {
	csrfToken, err := generateBrowserCSRFToken()
	if err != nil {
		return "", err
	}
	maxAge := 7 * 24 * 60 * 60
	if h != nil && h.cfg != nil && h.cfg.JWT.RefreshTokenExpireDays > 0 {
		maxAge = h.cfg.JWT.RefreshTokenExpireDays * 24 * 60 * 60
	}
	secure := h.authCookieSecure(c)
	http.SetCookie(c.Writer, &http.Cookie{
		Name: refreshTokenCookie, Value: refreshToken, Path: authCookiePath,
		MaxAge: maxAge, HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name: csrfTokenCookie, Value: csrfToken, Path: authCookiePath,
		MaxAge: maxAge, HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode,
	})
	return csrfToken, nil
}

func (h *AuthHandler) clearBrowserAuthCookies(c *gin.Context) {
	secure := h.authCookieSecure(c)
	for _, name := range []string{refreshTokenCookie, csrfTokenCookie} {
		http.SetCookie(c.Writer, &http.Cookie{
			Name: name, Value: "", Path: authCookiePath, MaxAge: -1,
			HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode,
		})
	}
}

func validateBrowserCSRF(c *gin.Context) error {
	if c == nil {
		return infraerrors.Forbidden("CSRF_INVALID", "invalid csrf token")
	}
	cookie, err := c.Cookie(csrfTokenCookie)
	if err != nil {
		return infraerrors.Forbidden("CSRF_INVALID", "invalid csrf token")
	}
	header := strings.TrimSpace(c.GetHeader(csrfTokenHeader))
	if header == "" || subtle.ConstantTimeCompare([]byte(cookie), []byte(header)) != 1 {
		return infraerrors.Forbidden("CSRF_INVALID", "invalid csrf token")
	}
	return nil
}

func generateBrowserCSRFToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func browserRefreshToken(c *gin.Context) (string, error) {
	if c == nil {
		return "", errors.New("missing request context")
	}
	value, err := c.Cookie(refreshTokenCookie)
	if err != nil || strings.TrimSpace(value) == "" {
		return "", infraerrors.Unauthorized("REFRESH_TOKEN_INVALID", "invalid refresh token")
	}
	return strings.TrimSpace(value), nil
}

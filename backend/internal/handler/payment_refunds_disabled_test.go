package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPaymentRefundEndpointsAreDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &PaymentHandler{}

	for _, tc := range []struct {
		name    string
		method  string
		path    string
		handler gin.HandlerFunc
	}{
		{name: "request", method: http.MethodPost, path: "/api/v1/payment/orders/1/refund-request", handler: h.RequestRefund},
		{name: "eligibility", method: http.MethodGet, path: "/api/v1/payment/orders/refund-eligible-providers", handler: h.GetRefundEligibleProviders},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(tc.method, tc.path, nil)

			tc.handler(c)

			require.Equal(t, http.StatusForbidden, recorder.Code)
			var payload response.Response
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
			require.Equal(t, "REFUNDS_DISABLED", payload.Reason)
		})
	}
}

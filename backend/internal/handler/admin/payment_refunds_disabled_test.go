package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAdminPaymentRefundEndpointsAreDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &PaymentHandler{}

	for _, tc := range []struct {
		name    string
		path    string
		handler gin.HandlerFunc
	}{
		{name: "process", path: "/api/v1/admin/payment/orders/1/refund", handler: h.ProcessRefund},
		{name: "query", path: "/api/v1/admin/payment/orders/1/refund/query", handler: h.QueryAndFinalizeRefund},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, tc.path, nil)

			tc.handler(c)

			require.Equal(t, http.StatusForbidden, recorder.Code)
			var payload response.Response
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
			require.Equal(t, "REFUNDS_DISABLED", payload.Reason)
		})
	}
}

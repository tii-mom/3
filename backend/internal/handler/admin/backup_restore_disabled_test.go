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

func TestBackupRestoreEndpointIsDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/backups/backup-id/restore", nil)

	(&BackupHandler{}).RestoreBackup(c)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	var payload response.Response
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	require.Equal(t, "ONLINE_RESTORE_DISABLED", payload.Reason)
}

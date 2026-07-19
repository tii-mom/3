package service

import (
	"context"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestProviderConfigurationRejectsRefundEnablement(t *testing.T) {
	svc := &PaymentConfigService{}

	_, err := svc.CreateProviderInstance(context.Background(), CreateProviderInstanceRequest{
		RefundEnabled: true,
	})
	require.Error(t, err)
	require.Equal(t, "REFUNDS_DISABLED", infraerrors.Reason(err))

	enabled := true
	_, err = svc.UpdateProviderInstance(context.Background(), 1, UpdateProviderInstanceRequest{
		AllowUserRefund: &enabled,
	})
	require.Error(t, err)
	require.Equal(t, "REFUNDS_DISABLED", infraerrors.Reason(err))
}

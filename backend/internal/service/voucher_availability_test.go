package service

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

type voucherSettingRepoStub struct {
	values map[string]string
}

func (s *voucherSettingRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *voucherSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (s *voucherSettingRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set call")
}

func (s *voucherSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			values[key] = value
		}
	}
	return values, nil
}

func (s *voucherSettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *voucherSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *voucherSettingRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

func voucherTestDecimal(t *testing.T, raw string) decimal.Decimal {
	t.Helper()
	value, err := decimal.NewFromString(raw)
	require.NoError(t, err)
	return value
}

func TestVoucherAvailabilityCalculatesServerAuthoritativeMaximum(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	settings := &voucherSettingRepoStub{values: map[string]string{
		"balance_voucher_enabled":       "true",
		"credit_bucket_enforce_enabled": "true",
		"balance_voucher_fee_bps":       "800",
		"balance_voucher_min_usd":       "10",
		"balance_voucher_max_usd":       "10000",
		"balance_voucher_daily_usd":     "30000",
		"balance_voucher_daily_count":   "10",
		"balance_voucher_expiry_days":   "30",
		"balance_voucher_step_up_usd":   "1000",
	}}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT transferable_credit::text, non_transferable_credit::text, debt::text FROM user_credit_accounts WHERE user_id = $1`)).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"transferable", "non_transferable", "debt"}).AddRow("108", "25", "0"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*), COALESCE(SUM(face_value), 0)::text FROM balance_vouchers WHERE tenant_id = $1 AND issuer_user_id = $2 AND created_at >= date_trunc('day', NOW())`)).
		WithArgs(int64(1), int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"count", "used"}).AddRow(2, "50"))

	result, err := NewVoucherService(db, settings, nil).Availability(context.Background(), 42)
	require.NoError(t, err)
	require.Equal(t, "108.00000000", result.TransferableCredit)
	require.Equal(t, "25.00000000", result.NonTransferableCredit)
	require.Equal(t, "100.00000000", result.MaximumFaceValueUSD)
	require.Equal(t, "29950.00000000", result.DailyRemainingUSD)
	require.Equal(t, int64(8), result.DailyRemainingCount)
	require.Equal(t, "1000.00000000", result.StepUpMinimumUSD)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMaximumVoucherFaceValueNeverExceedsTransferableCredit(t *testing.T) {
	transferable := voucherTestDecimal(t, "10.80")
	available := maximumVoucherFaceValue(transferable, 800)
	require.Equal(t, "10", available.String())
	require.False(t, available.Add(calculateVoucherFee(available, 800)).GreaterThan(transferable))
}

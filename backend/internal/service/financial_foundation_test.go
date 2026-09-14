package service

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestRechargeBaseMinorExcludesPaymentSurcharge(t *testing.T) {
	require.Equal(t, int64(10000), rechargeBaseMinor(decimal.RequireFromString("102.50"), decimal.RequireFromString("2.5")))
	require.Equal(t, int64(100000), rechargeBaseMinor(decimal.NewFromInt(1000), decimal.Zero))
	require.Zero(t, rechargeBaseMinor(decimal.NewFromInt(100), decimal.NewFromInt(-100)))
}

func TestFinancialOutboxRetryDelayIsBounded(t *testing.T) {
	require.Equal(t, time.Minute, financialOutboxRetryDelay(0))
	require.Equal(t, 8*time.Minute, financialOutboxRetryDelay(4))
	require.Equal(t, 128*time.Minute, financialOutboxRetryDelay(99))
}

func TestDistributionForecastRequiresRecentActivity(t *testing.T) {
	series := make([]DistributionAnalyticsPoint, 30)
	for index := range series {
		series[index].Date = time.Now().UTC().AddDate(0, 0, index-29).Format("2006-01-02")
	}
	forecast := forecastHorizon(series, 30, 14)
	require.False(t, forecast.Eligible)
	require.Equal(t, "insufficient_activity", forecast.Reason)
}

func TestDistributionForecastProjectsPositiveTrend(t *testing.T) {
	series := make([]DistributionAnalyticsPoint, 30)
	for index := range series {
		series[index] = DistributionAnalyticsPoint{
			Date:            time.Now().UTC().AddDate(0, 0, index-29).Format("2006-01-02"),
			SpendMinor:      int64(1000 + index*100),
			CommissionMinor: int64(100 + index*10),
		}
	}
	forecast := forecastHorizon(series, 7, 7)
	require.True(t, forecast.Eligible)
	require.Greater(t, forecast.EstimatedSpendMinor, int64(0))
	require.Greater(t, forecast.EstimatedCommissionMinor, int64(0))
	require.Greater(t, forecast.SpendGrowthPercent, float64(0))
	require.Greater(t, forecast.CommissionGrowthPercent, float64(0))
}

func commissionVector(base int64, rates [5]int64) []int64 {
	result := make([]int64, len(rates))
	for index, rate := range rates {
		result[index] = calculateCommissionMinor(base, rate)
	}
	return result
}

// TestCommissionVectorUsesPerLevelRates 只验证「按费率算佣金」这条纯函数，
// 与档位无关：推广计划已无档位，返佣比例来自商品上架时的设置。
func TestCommissionVectorUsesPerLevelRates(t *testing.T) {
	base := int64(100000)
	require.Equal(t, []int64{5000, 0, 0, 0, 0}, commissionVector(base, [5]int64{500, 0, 0, 0, 0}))
	require.Equal(t, []int64{8000, 800, 0, 0, 0}, commissionVector(base, [5]int64{800, 80, 0, 0, 0}))
}

func TestFirstRechargeBonusCapsRewardNotRecharge(t *testing.T) {
	require.True(t, calculateFirstRechargeBonus(decimal.NewFromInt(50000), 1000, decimal.NewFromInt(10000)).Equal(decimal.NewFromInt(5000)))
	require.True(t, calculateFirstRechargeBonus(decimal.NewFromInt(200000), 1000, decimal.NewFromInt(10000)).Equal(decimal.NewFromInt(10000)))
}

func TestWithdrawalFeeUsesMinorUnitsAndRoundsOnce(t *testing.T) {
	require.Equal(t, int64(0), calculateWithdrawalFee(10000, 0))
	require.Equal(t, int64(80), calculateWithdrawalFee(10000, 80))
	require.Equal(t, int64(1), calculateWithdrawalFee(101, 50))
}

// TestDistributionPolicyValidation 覆盖推广计划「钱包 / 提现 / 首充奖励」参数的校验。
// 推广计划已无档位：返佣比例不在这里配置，所以这里不再有任何 tiers 断言。
func TestDistributionPolicyValidation(t *testing.T) {
	input := DistributionPolicyInput{
		CommissionFreezeHours: 168, WithdrawalMinMinor: 2000, WithdrawalDailyLimit: 1,
		WithdrawalFeeBPS: 0, FirstRechargeBonusBPS: 1000, FirstRechargeBonusCap: "10000",
	}
	capAmount, err := validateDistributionPolicy(input)
	require.NoError(t, err)
	require.True(t, capAmount.Equal(decimal.NewFromInt(10000)))

	// 提现门槛必须为正。
	invalid := input
	invalid.WithdrawalMinMinor = 0
	_, err = validateDistributionPolicy(invalid)
	require.Error(t, err)

	// 提现手续费必须落在 [0, 10000) bps。
	invalid = input
	invalid.WithdrawalFeeBPS = 10000
	_, err = validateDistributionPolicy(invalid)
	require.Error(t, err)

	// 首充奖励比例必须落在 [0, 10000] bps。
	invalid = input
	invalid.FirstRechargeBonusBPS = 10001
	_, err = validateDistributionPolicy(invalid)
	require.Error(t, err)

	// 首充奖励上限必须是合法的非负金额。
	invalid = input
	invalid.FirstRechargeBonusCap = "not-a-number"
	_, err = validateDistributionPolicy(invalid)
	require.Error(t, err)
}

func TestVoucherFeeAndHash(t *testing.T) {
	require.True(t, calculateVoucherFee(decimal.RequireFromString("123.45678901"), 800).Equal(decimal.RequireFromString("9.87654312")))
	code, hash, last4, err := generateVoucherCode()
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(code, "VCH-"))
	digest := sha256.Sum256([]byte(code))
	require.Equal(t, hex.EncodeToString(digest[:]), hash)
	require.Len(t, last4, 4)
}

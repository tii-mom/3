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

func TestTierForVolumeUsesCurrentOrderBoundary(t *testing.T) {
	tiers := []DistributionTier{{Tier: 0, Threshold: 0}, {Tier: 1, Threshold: 100000}, {Tier: 2, Threshold: 1000000}, {Tier: 3, Threshold: 10000000}}
	require.Equal(t, 0, tierForVolume(tiers, 0))
	require.Equal(t, 0, tierForVolume(tiers, 99999))
	require.Equal(t, 1, tierForVolume(tiers, 100000))
	require.Equal(t, 1, tierForVolume(tiers, 999999))
	require.Equal(t, 2, tierForVolume(tiers, 1000000))
	require.Equal(t, 2, tierForVolume(tiers, 9999999))
	require.Equal(t, 3, tierForVolume(tiers, 10000000))
}

func TestCommissionTierUsesPreRechargeVolumeAtPromotionBoundary(t *testing.T) {
	tiers := []DistributionTier{
		{Tier: 0, Threshold: 0},
		{Tier: 1, Threshold: 100000},
		{Tier: 2, Threshold: 1000000},
		{Tier: 3, Threshold: 10000000},
	}
	// A recharge that moves a member from T1 to T2 must still use T1 for
	// that order; T2 applies to subsequent orders.
	require.Equal(t, 1, commissionTierForRecharge(tiers, 999999, nil))
	require.Equal(t, 2, tierForVolume(tiers, 1000000))
	override := 3
	require.Equal(t, 3, commissionTierForRecharge(tiers, 0, &override))
}

func TestComputeCompanyTierRatesCoverAllFiveDepartments(t *testing.T) {
	tiers := []DistributionTier{
		{Tier: 0, Threshold: 0, RatesBPS: [5]int64{1000, 0, 0, 0, 0}},
		{Tier: 1, Threshold: 100000, RatesBPS: [5]int64{1000, 400, 300, 200, 100}},
		{Tier: 2, Threshold: 1000000, RatesBPS: [5]int64{1500, 600, 400, 300, 200}},
		{Tier: 3, Threshold: 10000000, RatesBPS: [5]int64{2000, 800, 600, 400, 200}},
	}
	require.Equal(t, [5]int64{1000, 0, 0, 0, 0}, tiers[0].RatesBPS)
	require.Equal(t, [5]int64{1000, 400, 300, 200, 100}, tiers[1].RatesBPS)
	require.Equal(t, [5]int64{1500, 600, 400, 300, 200}, tiers[2].RatesBPS)
	require.Equal(t, [5]int64{2000, 800, 600, 400, 200}, tiers[3].RatesBPS)
	base := int64(100000)
	require.Equal(t, []int64{10000, 4000, 3000, 2000, 1000}, commissionVector(base, tiers[1].RatesBPS))
	require.Equal(t, []int64{15000, 6000, 4000, 3000, 2000}, commissionVector(base, tiers[2].RatesBPS))
	require.Equal(t, []int64{20000, 8000, 6000, 4000, 2000}, commissionVector(base, tiers[3].RatesBPS))
	require.Equal(t, []int64{10000, 0, 0, 0, 0}, commissionVector(base, tiers[0].RatesBPS))
}

func commissionVector(base int64, rates [5]int64) []int64 {
	result := make([]int64, len(rates))
	for index, rate := range rates {
		result[index] = calculateCommissionMinor(base, rate)
	}
	return result
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

func TestDistributionPolicyValidation(t *testing.T) {
	input := DistributionPolicyInput{
		CommissionFreezeHours: 168, WithdrawalMinMinor: 2000, WithdrawalDailyLimit: 1,
		WithdrawalFeeBPS: 0, FirstRechargeBonusBPS: 1000, FirstRechargeBonusCap: "10000",
		Tiers: []DistributionTier{
			{Tier: 0, Threshold: 0, RatesBPS: [5]int64{1000, 0, 0, 0, 0}},
			{Tier: 1, Threshold: 100000, RatesBPS: [5]int64{1000, 400, 300, 200, 100}},
			{Tier: 2, Threshold: 1000000, RatesBPS: [5]int64{1500, 600, 400, 300, 200}},
			{Tier: 3, Threshold: 10000000, RatesBPS: [5]int64{2000, 800, 600, 400, 200}},
		},
	}
	capAmount, err := validateDistributionPolicy(input)
	require.NoError(t, err)
	require.True(t, capAmount.Equal(decimal.NewFromInt(10000)))

	input.Tiers[2].Threshold = input.Tiers[1].Threshold
	_, err = validateDistributionPolicy(input)
	require.Error(t, err)

	input.Tiers[2].Threshold = 1000000
	input.Tiers[0].RatesBPS[1] = 1
	_, err = validateDistributionPolicy(input)
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

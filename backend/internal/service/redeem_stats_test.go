package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type redeemStatsRepositoryStub struct {
	RedeemCodeRepository
	stats *RedeemCodeStats
	now   time.Time
}

func (s *redeemStatsRepositoryStub) GetStats(_ context.Context, now time.Time) (*RedeemCodeStats, error) {
	s.now = now
	return s.stats, nil
}

func TestRedeemServiceGetStatsUsesRepositoryAggregation(t *testing.T) {
	expected := &RedeemCodeStats{
		TotalCodes:            10,
		ActiveCodes:           4,
		UsedCodes:             5,
		ExpiredCodes:          1,
		TotalValueDistributed: 30,
		ByType:                map[string]int64{RedeemTypeBalance: 10},
	}
	repo := &redeemStatsRepositoryStub{stats: expected}
	svc := NewRedeemService(repo, nil, nil, nil, nil, nil, nil, nil)

	got, err := svc.GetStats(context.Background())
	require.NoError(t, err)
	require.Same(t, expected, got)
	require.False(t, repo.now.IsZero())
}

func TestRedeemServiceGetStatsRejectsUnsupportedRepository(t *testing.T) {
	svc := NewRedeemService(struct{ RedeemCodeRepository }{}, nil, nil, nil, nil, nil, nil, nil)

	_, err := svc.GetStats(context.Background())
	require.ErrorContains(t, err, "statistics are unavailable")
}

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// 跨进程限流的语义契约：
//   1. 窗口内达到上限后拒绝，且被拒的请求不会把窗口往后延；
//   2. 窗口翻篇后计数重置；
//   3. 数据库异常时 fail-open（限流不该成为下单链路的单点故障）。

func newLimiterWithMock(t *testing.T) (*guestOrderLimiter, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return newGuestOrderLimiter(db), mock, func() { _ = db.Close() }
}

// expectWindowCheck 铺一次完整的「读改写」事务。
// storedHits 是上一轮遗留的计数，nextHits 是本次应当写回的值。
func expectWindowCheck(mock sqlmock.Sqlmock, storedStart time.Time, storedHits, nextHits int) {
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO shop_guest_order_windows").
		WithArgs("ip:1.2.3.4", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT window_start, hits FROM shop_guest_order_windows").
		WithArgs("ip:1.2.3.4").
		WillReturnRows(sqlmock.NewRows([]string{"window_start", "hits"}).AddRow(storedStart, storedHits))
	mock.ExpectExec("UPDATE shop_guest_order_windows").
		WithArgs("ip:1.2.3.4", sqlmock.AnyArg(), nextHits).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
}

func TestGuestOrderLimiterCountsOnlyAllowedAttempts(t *testing.T) {
	limiter, mock, done := newLimiterWithMock(t)
	defer done()
	ctx := context.Background()
	window := time.Now().UTC().Truncate(guestLimitWindow)

	// 首次：行是新插入的，hits=0 → 放行并写回 1
	expectWindowCheck(mock, window, 0, 1)
	require.True(t, limiter.allow(ctx, "ip:1.2.3.4", 2))

	// 第二次：hits=1 → 放行并写回 2
	expectWindowCheck(mock, window, 1, 2)
	require.True(t, limiter.allow(ctx, "ip:1.2.3.4", 2))

	// 第三次：已到上限 → 拒绝，且 hits 保持 2（不延长封禁窗口）
	expectWindowCheck(mock, window, 2, 2)
	require.False(t, limiter.allow(ctx, "ip:1.2.3.4", 2))

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGuestOrderLimiterResetsAfterWindowRolls(t *testing.T) {
	limiter, mock, done := newLimiterWithMock(t)
	defer done()

	// 库里存的是上一窗口的时间戳：本轮应当把计数重置为 1 并放行
	stale := time.Now().UTC().Truncate(guestLimitWindow).Add(-2 * guestLimitWindow)
	expectWindowCheck(mock, stale, 99, 1)
	require.True(t, limiter.allow(context.Background(), "ip:1.2.3.4", 2))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGuestOrderLimiterFailsOpenOnDatabaseError(t *testing.T) {
	limiter, mock, done := newLimiterWithMock(t)
	defer done()
	mock.ExpectBegin().WillReturnError(errors.New("db down"))

	require.True(t, limiter.allow(context.Background(), "ip:1.2.3.4", 2))
}

func TestGuestOrderLimiterFailsOpenOnUnexpectedRow(t *testing.T) {
	limiter, mock, done := newLimiterWithMock(t)
	defer done()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO shop_guest_order_windows").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT window_start, hits FROM shop_guest_order_windows").
		WillReturnRows(sqlmock.NewRows([]string{"window_start", "hits"}))
	mock.ExpectRollback()

	require.True(t, limiter.allow(context.Background(), "ip:1.2.3.4", 2))
}

func TestGuestOrderLimiterNilReceiversAreNoops(t *testing.T) {
	var limiter *guestOrderLimiter
	require.True(t, limiter.allow(context.Background(), "ip:1.2.3.4", 2))
	require.True(t, newGuestOrderLimiter(nil).allow(context.Background(), "ip:1.2.3.4", 2))
	// limit <= 0 视为未配置限流
	limiter2, _, done := newLimiterWithMock(t)
	defer done()
	require.True(t, limiter2.allow(context.Background(), "ip:1.2.3.4", 0))
}

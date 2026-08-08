package service

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestDistributionThawBindsProgramAndUserArguments(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company'`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectQuery(regexp.QuoteMeta(`
WITH distribution_thawed AS (
    UPDATE distribution_commissions
    SET status = 'AVAILABLE', thawed_at = NOW()
    WHERE program_id = $1 AND beneficiary_user_id = $2 AND status = 'FROZEN' AND frozen_until <= NOW()
    RETURNING amount_cny_minor
), shop_thawed AS (
    UPDATE shop_commission_records
    SET status = 'AVAILABLE', updated_at = NOW()
    WHERE tenant_id = 1 AND beneficiary_user_id = $2 AND status = 'FROZEN' AND frozen_until <= NOW()
    RETURNING amount_cny_minor
)
SELECT COALESCE((SELECT SUM(amount_cny_minor) FROM distribution_thawed), 0)
     + COALESCE((SELECT SUM(amount_cny_minor) FROM shop_thawed), 0)`)).
		WithArgs(int64(7), int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"amount"}).AddRow(0))
	mock.ExpectCommit()

	err = NewDistributionService(db, nil).Thaw(context.Background(), 42)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

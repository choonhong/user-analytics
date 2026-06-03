package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/choonhong/user-analytics/ent"
	"github.com/choonhong/user-analytics/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestLoginRepository(t *testing.T) {
	t.Run("rollups and idempotency", func(t *testing.T) {
		repo := newLoginRepository(t)
		ctx := context.Background()

		userID1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		userID2 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

		day1 := time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC)
		day1Later := time.Date(2026, 6, 3, 18, 0, 0, 0, time.UTC)
		day2 := time.Date(2026, 6, 4, 1, 0, 0, 0, time.UTC)

		require.NoError(t, repo.RecordLoginTx(ctx, userID1, day1))
		require.NoError(t, repo.RecordLoginTx(ctx, userID1, day1Later))
		require.NoError(t, repo.RecordLoginTx(ctx, userID2, day1))
		require.NoError(t, repo.RecordLoginTx(ctx, userID1, day2))

		require.NoError(t, repo.RecordLoginTx(ctx, userID1, day1))

		daily, err := repo.CountDailyUniqueUsers(ctx, "2026-06-03")
		require.NoError(t, err)
		require.Equal(t, 2, daily)

		dailyNext, err := repo.CountDailyUniqueUsers(ctx, "2026-06-04")
		require.NoError(t, err)
		require.Equal(t, 1, dailyNext)

		monthly, err := repo.CountMonthlyUniqueUsers(ctx, "2026-06")
		require.NoError(t, err)
		require.Equal(t, 2, monthly)
	})

	t.Run("month boundary", func(t *testing.T) {
		repo := newLoginRepository(t)
		ctx := context.Background()

		userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		require.NoError(t, repo.RecordLoginTx(ctx, userID, time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)))
		require.NoError(t, repo.RecordLoginTx(ctx, userID, time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)))

		jan, err := repo.CountDailyUniqueUsers(ctx, "2026-01-31")
		require.NoError(t, err)
		require.Equal(t, 1, jan)

		feb, err := repo.CountDailyUniqueUsers(ctx, "2026-02-01")
		require.NoError(t, err)
		require.Equal(t, 1, feb)

		janMonth, err := repo.CountMonthlyUniqueUsers(ctx, "2026-01")
		require.NoError(t, err)
		require.Equal(t, 1, janMonth)

		febMonth, err := repo.CountMonthlyUniqueUsers(ctx, "2026-02")
		require.NoError(t, err)
		require.Equal(t, 1, febMonth)
	})
}

func newLoginRepository(t *testing.T) *repository.LoginRepository {
	t.Helper()

	return repository.NewLoginRepository(openTestDB(t))
}

func openTestDB(t *testing.T) *ent.Client {
	t.Helper()

	name := strings.ReplaceAll(t.Name(), "/", "-")
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", name))
	require.NoError(t, err)

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	require.NoError(t, client.Schema.Create(context.Background()))

	t.Cleanup(func() {
		require.NoError(t, client.Close())
		require.NoError(t, db.Close())
	})

	return client
}

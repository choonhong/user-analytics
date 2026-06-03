package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/choonhong/user-analytics/internal/domain"
	"github.com/choonhong/user-analytics/internal/service"
	"github.com/choonhong/user-analytics/internal/service/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRecordLogin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		userID := uuid.New()
		timestamp := time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)

		svc, repo := newAnalyticsService(t)
		repo.EXPECT().RecordLoginTx(mock.Anything, userID, timestamp.UTC()).Return(nil).Once()

		require.NoError(t, svc.RecordLogin(context.Background(), userID, timestamp))
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		userID := uuid.New()
		timestamp := time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)

		svc, repo := newAnalyticsService(t)
		repo.EXPECT().RecordLoginTx(mock.Anything, userID, timestamp.UTC()).Return(errors.New("db down")).Once()

		require.Error(t, svc.RecordLogin(context.Background(), userID, timestamp))
	})
}

func TestGetDailyUserCount(t *testing.T) {
	t.Run("invalid date", func(t *testing.T) {
		t.Parallel()

		svc, _ := newAnalyticsService(t)

		_, err := svc.GetDailyUserCount(context.Background(), "bad-date")
		require.ErrorIs(t, err, domain.ErrInvalidDate)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		svc, repo := newAnalyticsService(t)
		repo.EXPECT().CountDailyUniqueUsers(mock.Anything, "2026-06-03").Return(5, nil).Once()

		count, err := svc.GetDailyUserCount(context.Background(), "2026-06-03")
		require.NoError(t, err)
		require.Equal(t, 5, count)
	})

	t.Run("repo error", func(t *testing.T) {
		t.Parallel()

		svc, repo := newAnalyticsService(t)
		repo.EXPECT().CountDailyUniqueUsers(mock.Anything, "2026-06-03").Return(0, errors.New("db down")).Once()

		_, err := svc.GetDailyUserCount(context.Background(), "2026-06-03")
		require.Error(t, err)
	})
}

func TestGetMonthlyUserCount(t *testing.T) {
	t.Run("invalid month", func(t *testing.T) {
		t.Parallel()

		svc, _ := newAnalyticsService(t)

		_, err := svc.GetMonthlyUserCount(context.Background(), "2026-13")
		require.ErrorIs(t, err, domain.ErrInvalidMonth)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		svc, repo := newAnalyticsService(t)
		repo.EXPECT().CountMonthlyUniqueUsers(mock.Anything, "2026-06").Return(10, nil).Once()

		count, err := svc.GetMonthlyUserCount(context.Background(), "2026-06")
		require.NoError(t, err)
		require.Equal(t, 10, count)
	})
}

func newAnalyticsService(t *testing.T) (*service.AnalyticsService, *mocks.MockLoginRepository) {
	t.Helper()
	repo := mocks.NewMockLoginRepository(t)

	return service.NewAnalyticsService(repo), repo
}

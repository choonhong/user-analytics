package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/choonhong/user-analytics/internal/api"
	"github.com/choonhong/user-analytics/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRecordLogin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		body := `{"user_id":"550e8400-e29b-41d4-a716-446655440000","login_time":"2026-06-03T12:00:00Z"}`
		req := httptest.NewRequest(http.MethodPost, "/v1/logins", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		api.NewRouter(newTestHandler(t)).ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("invalid user_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/logins", bytes.NewBufferString(`{"user_id":"not-a-uuid"}`))
		rec := httptest.NewRecorder()

		api.NewRouter(newTestHandler(t)).ServeHTTP(rec, req)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/logins", bytes.NewBufferString(`{`))
		rec := httptest.NewRecorder()

		api.NewRouter(newTestHandler(t)).ServeHTTP(rec, req)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestGetDailyUserCount(t *testing.T) {
	t.Run("missing date", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/daily-user-count", nil)
		rec := httptest.NewRecorder()

		api.NewRouter(newTestHandler(t)).ServeHTTP(rec, req)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid date", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/daily-user-count?date=bad-date", nil)
		rec := httptest.NewRecorder()

		api.NewRouter(newTestHandler(t)).ServeHTTP(rec, req)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/daily-user-count?date=2026-06-03", nil)
		rec := httptest.NewRecorder()

		api.NewRouter(newTestHandler(t)).ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		var count int
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&count))
		require.Equal(t, 3, count)
	})
}

func TestGetMonthlyUserCount(t *testing.T) {
	t.Run("missing month", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/monthly-user-count", nil)
		rec := httptest.NewRecorder()

		api.NewRouter(newTestHandler(t)).ServeHTTP(rec, req)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid month", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/monthly-user-count?month=2026-13", nil)
		rec := httptest.NewRecorder()

		api.NewRouter(newTestHandler(t)).ServeHTTP(rec, req)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/monthly-user-count?month=2026-06", nil)
		rec := httptest.NewRecorder()

		api.NewRouter(newTestHandler(t)).ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		var count int
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&count))
		require.Equal(t, 7, count)
	})
}

type stubRepo struct{}

func (stubRepo) RecordLoginTx(context.Context, uuid.UUID, time.Time) error { return nil }

func (stubRepo) CountDailyUniqueUsers(context.Context, string) (int, error) {
	return 3, nil
}

func (stubRepo) CountMonthlyUniqueUsers(context.Context, string) (int, error) {
	return 7, nil
}

var _ service.LoginRepository = stubRepo{}

func newTestHandler(t *testing.T) *api.Handler {
	t.Helper()

	return api.NewHandler(service.NewAnalyticsService(stubRepo{}))
}

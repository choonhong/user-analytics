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

type stubRepo struct{}

func (stubRepo) RecordLoginTx(context.Context, uuid.UUID, time.Time) error { return nil }
func (stubRepo) CountDailyUniqueUsers(context.Context, string) (int, error) {
	return 3, nil
}
func (stubRepo) CountMonthlyUniqueUsers(context.Context, string) (int, error) {
	return 7, nil
}

func newTestHandler() *api.Handler {
	return api.NewHandler(service.NewAnalyticsService(stubRepo{}))
}

func TestRecordLoginCreated(t *testing.T) {
	body := `{"user_id":"550e8400-e29b-41d4-a716-446655440000","login_time":"2026-06-03T12:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/logins", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	api.NewRouter(newTestHandler()).ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
}

func TestRecordLoginInvalidUUID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/logins", bytes.NewBufferString(`{"user_id":"not-a-uuid"}`))
	rec := httptest.NewRecorder()

	api.NewRouter(newTestHandler()).ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetDailyUserCount(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/daily-user-count?date=2026-06-03", nil)
	rec := httptest.NewRecorder()

	api.NewRouter(newTestHandler()).ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var count int
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&count))
	require.Equal(t, 3, count)
}

func TestGetMonthlyUserCount(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/monthly-user-count?month=2026-06", nil)
	rec := httptest.NewRecorder()

	api.NewRouter(newTestHandler()).ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var count int
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&count))
	require.Equal(t, 7, count)
}

func TestGetDailyUserCountMissingDate(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/daily-user-count", nil)
	rec := httptest.NewRecorder()

	api.NewRouter(newTestHandler()).ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

var _ service.LoginRepository = stubRepo{}

package api

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AnalyticsService interface {
	RecordLogin(ctx context.Context, userID uuid.UUID, loginTime time.Time) error
	GetDailyUserCount(ctx context.Context, date string) (int, error)
	GetMonthlyUserCount(ctx context.Context, month string) (int, error)
}

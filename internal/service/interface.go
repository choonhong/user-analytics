package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type LoginRepository interface {
	RecordLoginTx(ctx context.Context, userID uuid.UUID, loginTime time.Time) error
	CountDailyUniqueUsers(ctx context.Context, date string) (int, error)
	CountMonthlyUniqueUsers(ctx context.Context, month string) (int, error)
}

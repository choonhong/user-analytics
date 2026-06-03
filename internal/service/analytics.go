package service

import (
	"context"
	"time"

	"github.com/choonhong/user-analytics/internal/domain"
	"github.com/google/uuid"
)

type AnalyticsService struct {
	repo LoginRepository
}

func NewAnalyticsService(repo LoginRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) RecordLogin(ctx context.Context, userID uuid.UUID, loginTime time.Time) error {
	return s.repo.RecordLoginTx(ctx, userID, loginTime.UTC())
}

func (s *AnalyticsService) GetDailyUserCount(ctx context.Context, date string) (int, error) {
	if _, err := time.ParseInLocation(domain.DateLayout, date, time.UTC); err != nil {
		return 0, domain.ErrInvalidDate
	}

	return s.repo.CountDailyUniqueUsers(ctx, date)
}

func (s *AnalyticsService) GetMonthlyUserCount(ctx context.Context, month string) (int, error) {
	if _, err := time.ParseInLocation(domain.MonthLayout, month, time.UTC); err != nil {
		return 0, domain.ErrInvalidMonth
	}

	return s.repo.CountMonthlyUniqueUsers(ctx, month)
}

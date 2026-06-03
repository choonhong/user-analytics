package service

import (
	"context"
	"log"
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
	loginTime = loginTime.UTC()
	if err := s.repo.RecordLoginTx(ctx, userID, loginTime); err != nil {
		return err
	}

	log.Printf("Recorded login for user %s at %s.", userID, loginTime.Format(time.RFC3339))

	return nil
}

func (s *AnalyticsService) GetDailyUserCount(ctx context.Context, date string) (int, error) {
	if _, err := time.ParseInLocation(domain.DateLayout, date, time.UTC); err != nil {
		return 0, domain.ErrInvalidDate
	}

	count, err := s.repo.CountDailyUniqueUsers(ctx, date)
	if err != nil {
		return 0, err
	}

	log.Printf("Found %d unique users on %s.", count, date)

	return count, nil
}

func (s *AnalyticsService) GetMonthlyUserCount(ctx context.Context, month string) (int, error) {
	if _, err := time.ParseInLocation(domain.MonthLayout, month, time.UTC); err != nil {
		return 0, domain.ErrInvalidMonth
	}

	count, err := s.repo.CountMonthlyUniqueUsers(ctx, month)
	if err != nil {
		return 0, err
	}

	log.Printf("Found %d unique users in %s.", count, month)

	return count, nil
}

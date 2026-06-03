package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/choonhong/user-analytics/db/ent"
	"github.com/choonhong/user-analytics/db/ent/dailyuniqueuser"
	"github.com/choonhong/user-analytics/db/ent/monthlyuniqueuser"
	"github.com/choonhong/user-analytics/db/ent/userlogin"
	"github.com/choonhong/user-analytics/internal/domain"
	"github.com/google/uuid"
)

type LoginRepository struct {
	client *ent.Client
}

func NewLoginRepository(client *ent.Client) *LoginRepository {
	return &LoginRepository{client: client}
}

// RecordLoginTx records one login and updates rollups in a transaction.
func (r *LoginRepository) RecordLoginTx(ctx context.Context, userID uuid.UUID, loginTime time.Time) error {
	loginTime = loginTime.UTC()
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = insertLogin(ctx, tx, userID, loginTime); err != nil {
		return fmt.Errorf("record login (user_id=%s login_time=%s): %w", userID, loginTime, err)
	}

	if err = insertRollups(ctx, tx, userID, loginTime); err != nil {
		return fmt.Errorf("update rollups (user_id=%s login_time=%s): %w", userID, loginTime, err)
	}

	return tx.Commit()
}

func insertLogin(ctx context.Context, tx *ent.Tx, userID uuid.UUID, loginTime time.Time) error {
	return tx.UserLogin.Create().
		SetUserID(userID).
		SetLoginTime(loginTime).
		OnConflictColumns(userlogin.FieldUserID, userlogin.FieldLoginTime).
		Ignore().
		Exec(ctx)
}

func insertRollups(ctx context.Context, tx *ent.Tx, userID uuid.UUID, loginTime time.Time) error {
	loginTime = loginTime.UTC()
	date := loginTime.Format(domain.DateLayout)
	month := loginTime.Format(domain.MonthLayout)

	if err := tx.DailyUniqueUser.Create().
		SetDate(date).
		SetUserID(userID).
		OnConflictColumns(dailyuniqueuser.FieldDate, dailyuniqueuser.FieldUserID).
		Ignore().
		Exec(ctx); err != nil {
		return fmt.Errorf("insert daily unique user (date=%s user_id=%s): %w", date, userID, err)
	}

	if err := tx.MonthlyUniqueUser.Create().
		SetMonth(month).
		SetUserID(userID).
		OnConflictColumns(monthlyuniqueuser.FieldMonth, monthlyuniqueuser.FieldUserID).
		Ignore().
		Exec(ctx); err != nil {
		return fmt.Errorf("insert monthly unique user (month=%s user_id=%s): %w", month, userID, err)
	}

	return nil
}

// CountDailyUniqueUsers returns unique users for a UTC calendar date.
func (r *LoginRepository) CountDailyUniqueUsers(ctx context.Context, date string) (int, error) {
	return r.client.DailyUniqueUser.Query().
		Where(dailyuniqueuser.DateEQ(date)).
		Count(ctx)
}

// CountMonthlyUniqueUsers returns unique users for a UTC calendar month.
func (r *LoginRepository) CountMonthlyUniqueUsers(ctx context.Context, month string) (int, error) {
	return r.client.MonthlyUniqueUser.Query().
		Where(monthlyuniqueuser.MonthEQ(month)).
		Count(ctx)
}

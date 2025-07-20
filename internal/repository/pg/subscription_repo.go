package pg

import (
	"context"
	"fmt"
	"time"

	"github.com/Joshdike/subscriptions_aggregator/internal/models"
	"github.com/Joshdike/subscriptions_aggregator/internal/pkg/errors"
	"github.com/Joshdike/subscriptions_aggregator/internal/repository"
	"github.com/Joshdike/subscriptions_aggregator/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	sq "github.com/Masterminds/squirrel"
)

type SubscriptionRepo struct {
	pool *pgxpool.Pool
}

var _ repository.SubscriptionRepository = (*SubscriptionRepo)(nil)

func NewSubscriptionRepo(pool *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{
		pool: pool,
	}
}

func (s *SubscriptionRepo) Create(ctx context.Context, sub *models.SubscriptionRequest) (uint64, error) {
	startDate, err := utils.ParseMonthYear(sub.StartDate)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid start date", errors.ErrInvalidInput)
	}
	endDate := startDate.AddDate(0, 1, 0)

	if sub.EndDate != "" {
		endDate, err = utils.ParseMonthYear(sub.EndDate)
		if err != nil {
			return 0, fmt.Errorf("%w: invalid end date", errors.ErrInvalidInput)
		}
		if endDate.Before(startDate) {
			return 0, fmt.Errorf("%w: end date must be after start date", errors.ErrInvalidInput)
		}
	}

	subscription := models.RequestToSubscription(*sub, startDate, endDate)

	err = s.OverlapCheck(ctx, subscription)
	if err != nil {
		return 0, err
	}

	query, params, err := sq.Insert("subscriptions").
		Columns("service_name", "price", "user_id", "start_date", "end_date").
		Values(subscription.ServiceName, subscription.Price, subscription.UserID, subscription.StartDate, subscription.EndDate).
		Suffix("RETURNING id").PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return 0, fmt.Errorf("error creating query: %w", err)
	}

	var id uint64
	err = s.pool.QueryRow(ctx, query, params...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error creating subscription: %w", err)
	}

	return id, nil
}

func (s *SubscriptionRepo) GetAll(ctx context.Context) ([]models.SubscriptionResponse, error) {
	query, params, err := sq.Select("*").From("subscriptions").PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("error creating query: %w", err)
	}

	rows, err := s.pool.Query(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []models.SubscriptionResponse
	for rows.Next() {
		var sub models.Subscription
		err = rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate)
		if err != nil {
			return nil, fmt.Errorf("error scanning subscription: %w", err)
		}
		subscription := models.NewSubscriptionResponse(sub)
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, nil
}

func (s *SubscriptionRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.SubscriptionResponse, error) {
	query, params, err := sq.Select("*").From("subscriptions").Where("user_id = ?", userID).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("error creating query: %w", err)
	}

	rows, err := s.pool.Query(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []models.SubscriptionResponse
	for rows.Next() {
		var sub models.Subscription
		err = rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate)
		if err != nil {
			return nil, fmt.Errorf("error scanning subscription: %w", err)
		}
		subscription := models.NewSubscriptionResponse(sub)
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, nil
}

func (s *SubscriptionRepo) GetByID(ctx context.Context, id uint64) (models.SubscriptionResponse, error) {
	query, params, err := sq.Select("*").From("subscriptions").Where("id = ?", id).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return models.SubscriptionResponse{}, fmt.Errorf("error creating query: %w", err)
	}

	var sub models.Subscription
	err = s.pool.QueryRow(ctx, query, params...).Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.SubscriptionResponse{}, fmt.Errorf("%w: subscription not found", errors.ErrSubscriptionNotFound)
		}
		return models.SubscriptionResponse{}, fmt.Errorf("error getting subscription: %w", err)
	}

	subcription := models.NewSubscriptionResponse(sub)

	return subcription, nil
}

func (s *SubscriptionRepo) RenewOrExtend(ctx context.Context, id uint64) (uint64, error) {
	query, params, err := sq.Select(("*")).From("subscriptions").Where("id = ?", id).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return 0, fmt.Errorf("error creating query: %w", err)
	}

	var sub models.Subscription
	err = s.pool.QueryRow(ctx, query, params...).Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate)
	if err != nil {
		return 0, fmt.Errorf("error getting subscription: %w", err)
	}

	var newStartDate time.Time
	if sub.EndDate.After(time.Now()) {
		newStartDate = sub.EndDate
	} else {
		newStartDate = time.Now()
	}
	newSubscription := models.Subscription{
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   newStartDate,
		EndDate:     newStartDate.Add(sub.EndDate.Sub(sub.StartDate)),
	}

	query, params, err = sq.Insert("subscriptions").
		Columns("service_name", "price", "user_id", "start_date", "end_date").
		Values(newSubscription.ServiceName, newSubscription.Price, newSubscription.UserID, newSubscription.StartDate, newSubscription.EndDate).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).ToSql()

	if err != nil {
		return 0, fmt.Errorf("error creating query: %w", err)
	}

	var newId uint64
	err = s.pool.QueryRow(ctx, query, params...).Scan(&newId)
	if err != nil {
		return 0, fmt.Errorf("error getting subscription: %w", err)
	}
	return newId, nil
}

func (s *SubscriptionRepo) GetCost(ctx context.Context, userID uuid.UUID, serviceName string, start, end time.Time) (int, error) {
	query, params, err := sq.Select("SUM(price) AS total").From("subscriptions").
		Where("user_id = ?", userID).
		Where("service_name = ?", serviceName).
		Where("start_date >= ?", start).
		Where("end_date <= ?", end).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return 0, fmt.Errorf("error creating query: %w", err)
	}

	var totalcost int
	err = s.pool.QueryRow(ctx, query, params...).Scan(&totalcost)
	if err != nil {
		return 0, fmt.Errorf("error getting total: %w", err)
	}
	return totalcost, nil
}

func (s *SubscriptionRepo) Delete(ctx context.Context, id uint64) error {
	query, params, err := sq.Delete("subscriptions").Where("id = ?", id).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("error creating query: %w", err)
	}
	_, err = s.pool.Exec(ctx, query, params...)
	if err != nil {
		return fmt.Errorf("error deleting subscription: %w", err)
	}
	return nil
}

func (s *SubscriptionRepo) OverlapCheck(ctx context.Context, sub models.Subscription) error {
	query, params, err := sq.Select("COUNT(*)").
		From("subscriptions").
		Where("user_id = ?", sub.UserID).
		Where("service_name = ?", sub.ServiceName).
		Where("end_date > ?", sub.StartDate).
		Where("start_date < ?", sub.EndDate).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("error creating query: %w", err)
	}

	var count int
	err = s.pool.QueryRow(ctx, query, params...).Scan(&count)
	if err != nil {
		return fmt.Errorf("error checking overlap: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("%w: wait till current subscription ends or extend it", errors.ErrAlreadyExists)
	}

	return nil
}

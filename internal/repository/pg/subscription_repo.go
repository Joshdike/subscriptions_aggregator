package pg

import (
	"context"
	"fmt"
	"time"

	"github.com/Joshdike/subscriptions_aggregator/internal/models"
	"github.com/Joshdike/subscriptions_aggregator/internal/repository"
	"github.com/Joshdike/subscriptions_aggregator/internal/utils"
	"github.com/google/uuid"
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
		return 0, fmt.Errorf("invalid start date: %w", err)
	}
	endDate := startDate.AddDate(0, 1, 0)

	if sub.EndDate != "" {
		endDate, err = utils.ParseMonthYear(sub.EndDate)
		if err != nil {
			return 0, fmt.Errorf("invalid end date: %w", err)
		}
		if endDate.Before(startDate) {
			return 0, fmt.Errorf("end date must be after start date")
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

	var subscriptions []models.SubscriptionResponse
	for rows.Next() {
		var subscription models.SubscriptionResponse
		err = rows.Scan(&subscription.ID, &subscription.ServiceName, &subscription.Price, &subscription.UserID, &subscription.StartDate, &subscription.EndDate)
		if err != nil {
			return nil, fmt.Errorf("error scanning subscription: %w", err)
		}
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, nil
}

func (s *SubscriptionRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.SubscriptionResponse, error) {
	return nil, nil
}

func (s *SubscriptionRepo) GetByID(ctx context.Context, id uint64) (models.SubscriptionResponse, error) {
	return models.SubscriptionResponse{}, nil
}

func (s *SubscriptionRepo) Renew(ctx context.Context, id uint64) (uint64, error) {
	return 0, nil
}

func (s *SubscriptionRepo) GetCost(ctx context.Context, userID uuid.UUID, serviceName string, start, end time.Time) (int, error) {
	return 0, nil
}

func (s *SubscriptionRepo) Delete(ctx context.Context, id uint64) error {
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
		return fmt.Errorf("subscription already exists")
	}

	return nil
}

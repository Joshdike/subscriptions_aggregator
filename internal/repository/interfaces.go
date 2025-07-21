package repository

import (
	"context"
	"time"

	"github.com/Joshdike/subscriptions_aggregator/internal/models"
	"github.com/google/uuid"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *models.SubscriptionRequest) (uint64, error)
	GetAll(ctx context.Context) ([]models.AdminSubscriptionResponse, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.SubscriptionResponse, error)
	GetByID(ctx context.Context, id uint64) (models.SubscriptionResponse, error)
	Delete(ctx context.Context, id uint64) error
	RenewOrExtend(ctx context.Context, id uint64) (uint64, error)
	GetCost(ctx context.Context, userID uuid.UUID, serviceName string, start, end time.Time) (int, error)
	OverlapCheck(ctx context.Context, sub models.Subscription) (error)
}

package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uint64    `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
}

type SubscriptionRequest struct {
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date,omitempty"`
}

type SubscriptionResponse struct {
	ID          uint64    `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
}

func NewSubscriptionResponse(sub Subscription) SubscriptionResponse {
	return SubscriptionResponse{
		ID:          sub.ID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   sub.StartDate.Format("01-2006"),
		EndDate:     sub.EndDate.Format("01-2006"),
	}
}

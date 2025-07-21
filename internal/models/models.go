// Package models contains the core data structures for subscriptions
// Includes: 
//   - Database/domain models (Subscription)  
//   - API request/response DTOs  
//   - Conversion helpers between formats 
package models

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionRequest struct {
	ServiceName string    `json:"service_name"`  //name of the subscription service
	Price       int       `json:"price"` 		//price  in rubles (e.g. 400 = 400 rubles)
	UserID      uuid.UUID `json:"user_id"`		//uuid of the suscribing user
	StartDate   string    `json:"start_date"`	//Date in "MM-YYYY" format (e.g., "01-2025")
	EndDate     string    `json:"end_date,omitempty"`	// optional end date in "MM-YYYY" format
}

type Subscription struct {
	ID          uint64    `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Deleted     bool      `json:"deleted"` // Soft-delete flag (hidden from normal users)
}

type SubscriptionResponse struct {
	ID          uint64    `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
}

type AdminSubscriptionResponse struct {
	ID          uint64    `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
	Deleted     bool      `json:"deleted"`
}

// NewSubscriptionResponse converts Subscription(DB model) to API Response
//Formats date to "MM-YYYY"
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

func NewAdminSubscriptionResponse(sub Subscription) AdminSubscriptionResponse {
	return AdminSubscriptionResponse{
		ID:          sub.ID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   sub.StartDate.Format("01-2006"),
		EndDate:     sub.EndDate.Format("01-2006"),
		Deleted:     sub.Deleted,
	}
}

// RequestToSubscription converts SubscriptionRequest to Subscription(DB model)
//Requires pre-parsed start and end dates
func RequestToSubscription(sub SubscriptionRequest, start, end time.Time) Subscription {
	return Subscription{
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   start,
		EndDate:     end,
	}
}

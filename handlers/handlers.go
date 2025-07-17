package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type handle struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *handle {
	return &handle{db}
}

func (h *handle) CreateSubscription(http.ResponseWriter, *http.Request) {
	
}

func (h *handle) GetSubscriptions(http.ResponseWriter, *http.Request) {
	
}

func (h *handle) GetSubscriptionByID(http.ResponseWriter, *http.Request) {
	
}

func (h *handle) UpdateSubscription(http.ResponseWriter, *http.Request) {
	
}

func (h *handle) DeleteSubscription(http.ResponseWriter, *http.Request) {
	
}

func (h *handle) GetCostByDateRange(http.ResponseWriter, *http.Request) {
	
}
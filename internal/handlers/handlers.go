package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Joshdike/subscriptions_aggregator/internal/models"
	"github.com/Joshdike/subscriptions_aggregator/internal/utils"
	sq "github.com/Masterminds/squirrel"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type handle struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *handle {
	return &handle{db}
}

// Create a new subscription in the database
func (h *handle) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req models.SubscriptionRequest
	// Decode the JSON request body into the Subscription struct
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"error": "error decoding JSON"}`, http.StatusBadRequest)
		return
	}

	// parse the start date and end date from the request into time.Time
	StartDate, err := utils.ParseMonthYear(req.StartDate)
	if err != nil {
		http.Error(w, `{"error": "invalid start date"}`, http.StatusBadRequest)
		return
	}

	// if the end date is not provided, set it to the next month
	EndDate := StartDate.AddDate(0, 1, 0)

	// if the end date is provided, parse it and check if it is after the start date
	if req.EndDate != "" {
		EndDate, err = utils.ParseMonthYear(req.EndDate)
		if err != nil {
			http.Error(w, `{"error": "invalid end date"}`, http.StatusBadRequest)
			return
		}
		if EndDate.Before(StartDate) {
			http.Error(w, `{"error": "end date must be after start date"}`, http.StatusBadRequest)
			return
		}
	}
	subscription := models.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   StartDate,
		EndDate:     EndDate,
	}

	// generate sql query using squirrel
	query, params, err := sq.Insert("subscriptions").
		Columns("service_name", "price", "user_id", "start_date", "end_date").
		Values(subscription.ServiceName, subscription.Price, subscription.UserID, subscription.StartDate, subscription.EndDate).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).ToSql()

	if err != nil {
		http.Error(w, `{"error": "error creating query", "message": "internal sql error"}`, http.StatusInternalServerError)
		return
	}
	// insert the subscription into the database and get the id
	if err = h.db.QueryRow(r.Context(), query, params...).Scan(&subscription.ID); err != nil {
		http.Error(w, `{"error": "error inserting subscription"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]interface{}{"message": "subscription created successfully", "id": subscription.ID})
	if err != nil {
		http.Error(w, `{"error": "error encoding JSON"}`, http.StatusInternalServerError)
		return
	}

}

// Get all subscriptions from the database.
// only accessible by admin
func (h *handle) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// generate sql query using squirrel
	query, params, err := sq.Select(("*")).From("subscriptions").PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		http.Error(w, `{"error": "error creating query", "message": "internal sql error"}`, http.StatusInternalServerError)
		return
	}
	// get the subscriptions from the database
	rows, err := h.db.Query(r.Context(), query, params...)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error": "no subscriptions found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "error getting subscriptions"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var subscriptions []models.SubscriptionResponse
	for rows.Next() {
		var sub models.Subscription
		if err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate); err != nil {
			http.Error(w, `{"error": "error getting subscriptions"}`, http.StatusInternalServerError)
			return
		}
		// map the subscription to the SubscriptionResponse struct
		subscription := models.NewSubscriptionResponse(sub)
		// append the subscription to the subscriptions slice
		subscriptions = append(subscriptions, subscription)
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(subscriptions)
	if err != nil {
		http.Error(w, `{"error": "error encoding JSON"}`, http.StatusInternalServerError)
		return
	}

}

// Get a single subscription from the database by id
func (h *handle) GetSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the id from the url and validate it
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error": "invalid subscription id"}`, http.StatusBadRequest)
		return
	}

	// generate sql query using squirrel
	query, params, err := sq.Select(("*")).From("subscriptions").Where("id = ?", id).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		http.Error(w, `{"error": "error creating query", "message": "internal sql error"}`, http.StatusInternalServerError)
		return
	}

	var subscription models.Subscription
	// get the subscription from the database
	err = h.db.QueryRow(r.Context(), query, params...).Scan(&subscription.ID, &subscription.ServiceName, &subscription.Price, &subscription.UserID, &subscription.StartDate, &subscription.EndDate)
	if err != nil {
		http.Error(w, `{"error": "error getting subscription"}`, http.StatusInternalServerError)
		return
	}

	// map the subscription to the SubscriptionResponse struct
	resp := models.NewSubscriptionResponse(subscription)

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, `{"error": "error encoding JSON"}`, http.StatusInternalServerError)
		return
	}

}

// Get all subscriptions from the database by user id
func (h *handle) GetSubscriptionByUserID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the user id from the url and validate it
	user_id, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		http.Error(w, `{"error": "invalid user id"}`, http.StatusBadRequest)
		return
	}

	// generate sql query using squirrel
	query, params, err := sq.Select(("*")).From("subscriptions").Where("user_id = ?", user_id).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		http.Error(w, `{"error": "error creating query", "message": "internal sql error"}`, http.StatusInternalServerError)
		return
	}
	// get the subscriptions from the database
	rows, err := h.db.Query(r.Context(), query, params...)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error": "no subscriptions found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "error getting subscriptions"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var subscriptions []models.SubscriptionResponse
	for rows.Next() {
		var subscription models.Subscription
		if err := rows.Scan(&subscription.ID, &subscription.ServiceName, &subscription.Price, &subscription.UserID, &subscription.StartDate, &subscription.EndDate); err != nil {
			http.Error(w, `{"error": "error getting subscriptions"}`, http.StatusInternalServerError)
			return
		}
		// map the subscription to the SubscriptionResponse struct
		resp := models.NewSubscriptionResponse(subscription)
		// append the subscription to the subscriptions slice
		subscriptions = append(subscriptions, resp)
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(subscriptions)
	if err != nil {
		http.Error(w, `{"error": "error encoding JSON"}`, http.StatusInternalServerError)
		return
	}
}

// Renew a subscription
func (h *handle) RenewSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the id from the url and validate it
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error": "invalid subscription id"}`, http.StatusBadRequest)
		return
	}

	// generate sql query using squirrel
	query, params, err := sq.Select(("*")).From("subscriptions").Where("id = ?", id).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		http.Error(w, `{"error": "error creating query", "message": "internal sql error"}`, http.StatusInternalServerError)
		return
	}

	// get the previous subscription from the database
	var subscription models.Subscription
	err = h.db.QueryRow(r.Context(), query, params...).Scan(&subscription.ID, &subscription.ServiceName, &subscription.Price, &subscription.UserID, &subscription.StartDate, &subscription.EndDate)
	if err != nil {
		http.Error(w, `{"error": "error getting subscription"}`, http.StatusInternalServerError)
		return
	}

	// check if the subscription is expired before renewing it and set the new start date
	var newStartDate time.Time
	if subscription.EndDate.After(time.Now()) {
		newStartDate = subscription.EndDate
	} else {
		newStartDate = time.Now()
	}

	// create a new subscription
	newSubscription := models.Subscription{
		ServiceName: subscription.ServiceName,
		Price:       subscription.Price,
		UserID:      subscription.UserID,
		StartDate:   newStartDate,
		EndDate:     newStartDate.Add(subscription.EndDate.Sub(subscription.StartDate)),
	}

	// generate sql query using squirrel
	query, params, err = sq.Insert("subscriptions").
		Columns("service_name", "price", "user_id", "start_date", "end_date").
		Values(newSubscription.ServiceName, newSubscription.Price, newSubscription.UserID, newSubscription.StartDate, newSubscription.EndDate).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).ToSql()

	if err != nil {
		http.Error(w, `{"error": "error creating query", "message": "internal sql error"}`, http.StatusInternalServerError)
		return
	}

	//insert the new subscription and get the new subscription id
	err = h.db.QueryRow(r.Context(), query, params...).Scan(&newSubscription.ID)
	if err != nil {
		http.Error(w, `{"error": "error creating subscription"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]interface{}{"message": "subscription renewed successfully", "new_id": newSubscription.ID})
	if err != nil {
		http.Error(w, `{"error": "error encoding JSON"}`, http.StatusInternalServerError)
		return
	}

}

func (h *handle) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the id from the url and validate it
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error": "invalid subscription id"}`, http.StatusBadRequest)
		return
	}

	// generate sql query using squirrel
	query, params, err := sq.Delete("subscriptions").Where("id = ?", id).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		http.Error(w, `{"error": "error creating query", "message": "internal sql error"}`, http.StatusInternalServerError)
		return
	}

	// delete the subscription from the database
	_, err = h.db.Exec(r.Context(), query, params...)
	if err != nil {
		http.Error(w, `{"error": "error deleting subscription"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]interface{}{"message": "subscription deleted successfully"})
	if err != nil {
		http.Error(w, `{"error": "error encoding JSON"}`, http.StatusInternalServerError)
		return
	}
}

func (h *handle) GetCostByDateRange(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the user id from the url and validate it
	user_id, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		http.Error(w, `{"error": "invalid user id"}`, http.StatusBadRequest)
		return
	}
	//get the service name from the url and validate it
	service_name := r.URL.Query().Get("service_name")
	if service_name == "" {
		http.Error(w, `{"error": "invalid service name"}`, http.StatusBadRequest)
		return
	}

	// get the start date and end date from the url and validate them
	startDate, err := utils.ParseMonthYear(r.URL.Query().Get("from"))
	if err != nil {
		http.Error(w, `{"error": "invalid start date"}`, http.StatusBadRequest)
		return
	}
	endDate, err := utils.ParseMonthYear(r.URL.Query().Get("to"))
	if err != nil {
		http.Error(w, `{"error": "invalid end date"}`, http.StatusBadRequest)
		return
	}

	// generate sql query using squirrel
	query, params, err := sq.Select(("SUM(price) AS total")).From("subscriptions").Where("user_id = ? AND service_name = ? AND start_date >= ? AND end_date <= ?", user_id, service_name, startDate, endDate).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		http.Error(w, `{"error": "error creating query", "message": "internal sql error"}`, http.StatusInternalServerError)
		return
	}

	// get the cost from the database
	var totalcost int
	err = h.db.QueryRow(r.Context(), query, params...).Scan(&totalcost)
	if err != nil {
		http.Error(w, `{"error": "error getting cost"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]interface{}{"total cost": totalcost})
	if err != nil {
		http.Error(w, `{"error": "error encoding JSON"}`, http.StatusInternalServerError)
		return
	}
}

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Joshdike/subscriptions_aggregator/internal/models"
	"github.com/Joshdike/subscriptions_aggregator/internal/repository"
	"github.com/Joshdike/subscriptions_aggregator/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	repo repository.SubscriptionRepository
}

func New(repo repository.SubscriptionRepository) *SubscriptionHandler {
	return &SubscriptionHandler{repo: repo}
}

// Create a new subscription in the database
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req models.SubscriptionRequest
	// Decode the JSON request body into the Subscription struct
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"error": "error decoding JSON"}`, http.StatusBadRequest)
		return
	}

	// Call the Create function in the repository to create the subscription in the database
	id, err := h.repo.Create(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error": "internal sql error", "description": "error creating subscription"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]interface{}{"message": "subscription created successfully", "id": id})
	if err != nil {
		http.Error(w, `{"error": "error encoding JSON"}`, http.StatusInternalServerError)
		return
	}

}

// Get all subscriptions from the database.
// only accessible by admin
func (h *SubscriptionHandler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the subscriptions from the database
	subscriptions, err := h.repo.GetAll(r.Context())
	if err != nil {
		http.Error(w, `{"error": "internal sql error", "description": "error getting subscriptions"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(subscriptions)
	if err != nil {
		http.Error(w, `{"error": "error encoding JSON"}`, http.StatusInternalServerError)
		return
	}

}

// Get a single subscription from the database by id
func (h *SubscriptionHandler) GetSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the id from the url and validate it
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error": "invalid subscription id"}`, http.StatusBadRequest)
		return
	}

	// get the subscription from the database
	subscription, err := h.repo.GetByID(r.Context(), uint64(id))
	if err != nil {
		http.Error(w, `{"error": "internal sql error", "description": "error getting subscription"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(subscription)
	if err != nil {
		http.Error(w, `{"error": "error encoding JSON"}`, http.StatusInternalServerError)
		return
	}

}

// Get all subscriptions from the database by user id
func (h *SubscriptionHandler) GetSubscriptionByUserID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the user id from the url and validate it
	user_id, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		http.Error(w, `{"error": "invalid user id"}`, http.StatusBadRequest)
		return
	}

	// get the subscriptions from the database
	subscriptions, err := h.repo.GetByUserID(r.Context(), user_id)
	if err != nil {
		http.Error(w, `{"error": "internal sql error", "description": "error getting subscriptions"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(subscriptions)
	if err != nil {
		http.Error(w, `{"error": "error encoding JSON"}`, http.StatusInternalServerError)
		return
	}
}

// Renew a subscription
func (h *SubscriptionHandler) RenewOrExtendSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the id from the url and validate it
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error": "invalid subscription id"}`, http.StatusBadRequest)
		return
	}

	// renew the subscription in the database
	newId, err := h.repo.RenewOrExtend(r.Context(), uint64(id))
	if err != nil {
		http.Error(w, `{"error": "internal sql error", "description": "error renewing subscription"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]interface{}{"message": "subscription renewed successfully", "new_id": newId})
	if err != nil {
		http.Error(w, `{"error": "error encoding JSON"}`, http.StatusInternalServerError)
		return
	}

}

func (h *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the id from the url and validate it
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error": "invalid subscription id"}`, http.StatusBadRequest)
		return
	}

	// delete the subscription from the database
	err = h.repo.Delete(r.Context(), uint64(id))
	if err != nil {
		http.Error(w, `{"error": "internal sql error", "description": "error deleting subscription"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]interface{}{"message": "subscription deleted successfully"})
	if err != nil {
		http.Error(w, `{"error": "error encoding JSON"}`, http.StatusInternalServerError)
		return
	}
}

func (h *SubscriptionHandler) GetCostByDateRange(w http.ResponseWriter, r *http.Request) {
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

	// get the cost from the database
	totalcost, err := h.repo.GetCost(r.Context(), user_id, service_name, startDate, endDate)
	if err != nil {
		http.Error(w, `{"error": "internal sql error", "description": "error getting cost"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]interface{}{"total cost": totalcost})
	if err != nil {
		http.Error(w, `{"error": "error encoding JSON"}`, http.StatusInternalServerError)
		return
	}
}

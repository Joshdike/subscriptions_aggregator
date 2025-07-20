package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Joshdike/subscriptions_aggregator/internal/models"
	"github.com/Joshdike/subscriptions_aggregator/internal/pkg/errors"
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
		err = errors.ErrDecodingJSON
		utils.WriteError(w, err)
		return
	}

	// Call the Create function in the repository to create the subscription in the database
	id, err := h.repo.Create(r.Context(), &req)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]interface{}{"message": "subscription created successfully", "id": id})
	if err != nil {
		err = errors.ErrEncodingJSON
		utils.WriteError(w, err)
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
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(subscriptions)
	if err != nil {
		err = errors.ErrEncodingJSON
		utils.WriteError(w, err)
		return
	}

}

// Get a single subscription from the database by id
func (h *SubscriptionHandler) GetSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the id from the url and validate it
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		err = fmt.Errorf("%w: invalid subscription id", errors.ErrInvalidInput)
		utils.WriteError(w, err)
		return
	}

	// get the subscription from the database
	subscription, err := h.repo.GetByID(r.Context(), uint64(id))
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(subscription)
	if err != nil {
		err = errors.ErrEncodingJSON
		utils.WriteError(w, err)
		return
	}

}

// Get all subscriptions from the database by user id
func (h *SubscriptionHandler) GetSubscriptionByUserID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the user id from the url and validate it
	user_id, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		err = fmt.Errorf("%w: invalid user id", errors.ErrInvalidInput)
		utils.WriteError(w, err)
		return
	}

	// get the subscriptions from the database
	subscriptions, err := h.repo.GetByUserID(r.Context(), user_id)
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(subscriptions)
	if err != nil {
		err = errors.ErrEncodingJSON
		utils.WriteError(w, err)
		return
	}
}

// Renew a subscription
func (h *SubscriptionHandler) RenewOrExtendSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the id from the url and validate it
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		err = fmt.Errorf("%w: invalid subscription id", errors.ErrInvalidInput)
		utils.WriteError(w, err)
		return
	}

	// renew the subscription in the database
	newId, err := h.repo.RenewOrExtend(r.Context(), uint64(id))
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]interface{}{"message": "subscription renewed successfully", "new_id": newId})
	if err != nil {
		err = errors.ErrEncodingJSON
		utils.WriteError(w, err)
		return
	}

}

func (h *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the id from the url and validate it
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		err = fmt.Errorf("%w: invalid subscription id", errors.ErrInvalidInput)
		utils.WriteError(w, err)
		return
	}

	// delete the subscription from the database
	err = h.repo.Delete(r.Context(), uint64(id))
	if err != nil {
		utils.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]interface{}{"message": "subscription deleted successfully"})
	if err != nil {
		err = errors.ErrEncodingJSON
		utils.WriteError(w, err)
		return
	}
}

func (h *SubscriptionHandler) GetCostByDateRange(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the user id from the url and validate it
	user_id, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		err = fmt.Errorf("%w: invalid user id", errors.ErrInvalidInput)
		utils.WriteError(w, err)
		return
	}
	//get the service name from the url and validate it
	service_name := r.URL.Query().Get("service_name")
	if service_name == "" {
		err = fmt.Errorf("%w: invalid service name", errors.ErrInvalidInput)
		utils.WriteError(w, err)
		return
	}

	// get the start date and end date from the url and validate them
	startDate, err := utils.ParseMonthYear(r.URL.Query().Get("from"))
	if err != nil {
		err = fmt.Errorf("%w: invalid start date, %s", errors.ErrInvalidInput, err.Error())
		utils.WriteError(w, err)
		return
	}
	endDate, err := utils.ParseMonthYear(r.URL.Query().Get("to"))
	if err != nil {
		err = fmt.Errorf("%w: invalid end date, %s", errors.ErrInvalidInput, err.Error())
		utils.WriteError(w, err)
		return
	}

	// get the cost from the database
	totalcost, err := h.repo.GetCost(r.Context(), user_id, service_name, startDate, endDate)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]interface{}{"total cost": totalcost})
	if err != nil {
		err = errors.ErrEncodingJSON
		utils.WriteError(w, err)
		return
	}
}

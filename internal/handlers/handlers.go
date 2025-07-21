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

// CreateSubscription godoc
// @Summary Create a new subscription
// @Description Creates a new subscription for a user
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body models.SubscriptionRequest true "Subscription creation data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req models.SubscriptionRequest
	
	//Decode the request body and validate
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		err = errors.ErrDecodingJSON
		utils.WriteError(w, err)
		return
	}

	//Create the subscription
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

// GetSubscriptions godoc
// @Summary Get all subscriptions (Admin Only)
// @Description Retrieves complete list of all subscriptions. Requires admin privileges.
// @Tags admin
// @Produce json
// @Security ApiKeyAuth
// @Param X-Admin-Secret header string true "Admin secret key"
// @Success 200 {array} models.AdminSubscriptionResponse
// @Failure 401 {object} utils.ErrorResponse 
// @Failure 500 {object} utils.ErrorResponse 
// @Router /subscriptions [get]
func (h *SubscriptionHandler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//Get all subscriptions by admin
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

// GetSubscriptionByID godoc
// @Summary Get a specific subscription by ID
// @Description Retrieves a specific subscription by its numeric ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} models.SubscriptionResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the subscription ID from the URL, convert to int and validate it
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		err = fmt.Errorf("%w: invalid subscription id", errors.ErrInvalidInput)
		utils.WriteError(w, err)
		return
	}

	// Get the subscription
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

// GetSubscriptionByUserID godoc
// @Summary Get all subscriptions for a user
// @Description Retrieves all subscriptions for a specific user
// @Tags subscriptions
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {array} models.SubscriptionResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /subscriptions/user/{user_id} [get]
func (h *SubscriptionHandler) GetSubscriptionByUserID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the user ID from the URL and validate it
	user_id, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		err = fmt.Errorf("%w: invalid user id", errors.ErrInvalidInput)
		utils.WriteError(w, err)
		return
	}
	// Get the subscriptions
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

// RenewOrExtendSubscription godoc
// @Summary Renew or extend a subscription
// @Description Renews or extends an existing subscription
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /subscriptions/{id} [post]
func (h *SubscriptionHandler) RenewOrExtendSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Get the subscription id from the URL and validate it
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		err = fmt.Errorf("%w: invalid subscription id", errors.ErrInvalidInput)
		utils.WriteError(w, err)
		return
	}

	// Renew or extend the subscription and get the new id
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

// DeleteSubscription godoc
// @Summary Soft delete a subscription
// @Description Marks a subscription as deleted by setting 'deleted' flag to true (does not permanently remove)
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription ID" minimum(1)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /subscriptions/{id} [patch]
func (h *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the id from the url and validate it
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		err = fmt.Errorf("%w: invalid subscription id", errors.ErrInvalidInput)
		utils.WriteError(w, err)
		return
	}
	//soft delete the subscription
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

// GetCostByDateRange godoc
// @Summary Get the cost of a subscription for a specific date range
// @Description Retrieves the cost of a subscription for a specific date range
// @Tags subscriptions
// @Produce json
// @Param user_id path string true "User ID"
// @Param service_name query string true "Service Name"
// @Param from query string true "Start Date"
// @Param to query string true "End Date"
// @Success 200 {object} map[string]int
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /costs/{user_id} [get]
func (h *SubscriptionHandler) GetCostByDateRange(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the user id from the url and validate it
	user_id, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		err = fmt.Errorf("%w: invalid user id", errors.ErrInvalidInput)
		utils.WriteError(w, err)
		return
	}
	
	// get the service name from the query and validate it
	service_name := r.URL.Query().Get("service_name")
	if service_name == "" {
		err = fmt.Errorf("%w: invalid service name", errors.ErrInvalidInput)
		utils.WriteError(w, err)
		return
	}

	// get the start date, parse and validate it
	startDate, err := utils.ParseMonthYear(r.URL.Query().Get("from"))
	if err != nil {
		err = fmt.Errorf("%w: invalid start date, %s", errors.ErrInvalidInput, err.Error())
		utils.WriteError(w, err)
		return
	}

	// get the end date, parse and validate it
	endDate, err := utils.ParseMonthYear(r.URL.Query().Get("to"))
	if err != nil {
		err = fmt.Errorf("%w: invalid end date, %s", errors.ErrInvalidInput, err.Error())
		utils.WriteError(w, err)
		return
	}

	// get the cost
	totalcost, err := h.repo.GetCost(r.Context(), user_id, service_name, startDate, endDate)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]int{"total cost": totalcost})
	if err != nil {
		err = errors.ErrEncodingJSON
		utils.WriteError(w, err)
		return
	}
}

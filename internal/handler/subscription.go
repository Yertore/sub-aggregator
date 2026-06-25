package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Yertore/sub-aggregator/internal/apperror"
	"github.com/Yertore/sub-aggregator/internal/model"
)

// Service defines the interface that the handler depends on.
type Service interface {
	Create(ctx context.Context, req *model.CreateSubscriptionRequest) (*model.Subscription, error)
	GetByID(ctx context.Context, id string) (*model.Subscription, error)
	List(ctx context.Context, userID, serviceName string, limit, offset int) ([]*model.Subscription, error)
	Update(ctx context.Context, id string, req *model.UpdateSubscriptionRequest) (*model.Subscription, error)
	Delete(ctx context.Context, id string) error
	TotalCost(ctx context.Context, userID, serviceName, from, to string) (int, error)
}

type Handler struct {
	svc Service
}

func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Create godoc
// @Summary      Создать подписку
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        request body model.CreateSubscriptionRequest true "Данные подписки"
// @Success      201 {object} model.Subscription
// @Failure      400 {object} map[string]string
// @Router       /subscriptions [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sub, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, sub)
}

// GetByID godoc
// @Summary      Получить подписку по ID
// @Tags         subscriptions
// @Produce      json
// @Param        id path string true "UUID подписки"
// @Success      200 {object} model.Subscription
// @Failure      404 {object} map[string]string
// @Router       /subscriptions/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	sub, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, sub)
}

// List godoc
// @Summary      Список подписок
// @Tags         subscriptions
// @Produce      json
// @Param        user_id      query string false "UUID пользователя"
// @Param        service_name query string false "Название сервиса"
// @Param        limit        query int    false "Лимит (по умолчанию 20, макс 100)"
// @Param        offset       query int    false "Смещение (по умолчанию 0)"
// @Success      200 {array}  model.Subscription
// @Failure      500 {object} map[string]string
// @Router       /subscriptions [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	subs, err := h.svc.List(r.Context(), userID, serviceName, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if subs == nil {
		subs = []*model.Subscription{}
	}

	writeJSON(w, http.StatusOK, subs)
}

// Update godoc
// @Summary      Обновить подписку
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id      path string true "UUID подписки"
// @Param        request body model.UpdateSubscriptionRequest true "Поля для обновления"
// @Success      200 {object} model.Subscription
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /subscriptions/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req model.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sub, err := h.svc.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, sub)
}

// Delete godoc
// @Summary      Удалить подписку
// @Tags         subscriptions
// @Param        id path string true "UUID подписки"
// @Success      204
// @Failure      404 {object} map[string]string
// @Router       /subscriptions/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TotalCost godoc
// @Summary      Суммарная стоимость подписок
// @Tags         subscriptions
// @Produce      json
// @Param        user_id      query string false "UUID пользователя"
// @Param        service_name query string false "Название сервиса"
// @Param        from         query string false "Начало периода MM-YYYY"
// @Param        to           query string false "Конец периода MM-YYYY"
// @Success      200 {object} map[string]int
// @Failure      400 {object} map[string]string
// @Router       /subscriptions/cost [get]
func (h *Handler) TotalCost(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	total, err := h.svc.TotalCost(r.Context(), userID, serviceName, from, to)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"total_cost": total})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

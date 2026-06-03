package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/choonhong/user-analytics/internal/domain"
	"github.com/choonhong/user-analytics/internal/service"
	"github.com/google/uuid"
)

type Handler struct {
	svc *service.AnalyticsService
}

func NewHandler(svc *service.AnalyticsService) *Handler {
	return &Handler{svc: svc}
}

type loginRequest struct {
	UserID    string `json:"user_id"`
	LoginTime string `json:"login_time,omitempty"`
}

func (h *Handler) RecordLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")

		return
	}

	userID, loginTime, err := parseLoginRequest(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())

		return
	}

	if err := h.svc.RecordLogin(r.Context(), userID, loginTime); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to record login")

		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetDailyUserCount(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if err := validateDateParam(date); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())

		return
	}

	count, err := h.svc.GetDailyUserCount(r.Context(), date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get daily count")

		return
	}

	writeJSON(w, http.StatusOK, count)
}

func (h *Handler) GetMonthlyUserCount(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if err := validateMonthParam(month); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())

		return
	}

	count, err := h.svc.GetMonthlyUserCount(r.Context(), month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get monthly count")

		return
	}

	writeJSON(w, http.StatusOK, count)
}

func parseLoginRequest(req loginRequest) (uuid.UUID, time.Time, error) {
	if strings.TrimSpace(req.UserID) == "" {
		return uuid.Nil, time.Time{}, errors.New("user_id is required")
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return uuid.Nil, time.Time{}, errors.New("invalid user_id")
	}

	var loginTime time.Time
	if req.LoginTime == "" {
		loginTime = time.Now().UTC()
	} else {
		loginTime, err = time.Parse(time.RFC3339, req.LoginTime)
		if err != nil {
			return uuid.Nil, time.Time{}, errors.New("invalid login_time, expected RFC3339")
		}
		loginTime = loginTime.UTC()
	}

	return userID, loginTime, nil
}

func validateDateParam(date string) error {
	if date == "" {
		return errors.New("date query parameter is required")
	}
	if _, err := time.ParseInLocation(domain.DateLayout, date, time.UTC); err != nil {
		return domain.ErrInvalidDate
	}

	return nil
}

func validateMonthParam(month string) error {
	if month == "" {
		return errors.New("month query parameter is required")
	}
	if _, err := time.ParseInLocation(domain.MonthLayout, month, time.UTC); err != nil {
		return domain.ErrInvalidMonth
	}

	return nil
}

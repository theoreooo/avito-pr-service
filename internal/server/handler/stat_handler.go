package handler

import (
	"net/http"

	"avito-pr-service/internal/service"
)

type StatisticsHandler struct {
	service *service.StatisticsService
}

func NewStatisticsHandler(service *service.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{service: service}
}

func (h *StatisticsHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStatistics(r.Context())
	if err != nil {
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}

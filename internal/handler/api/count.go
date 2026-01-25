package api

import (
	"encoding/json"
	"net/http"

	"topspots-backend/internal/model"
	"topspots-backend/internal/service"
)

type CountHandler struct {
	service *service.PlacesService
}

func NewCountHandler(svc *service.PlacesService) *CountHandler {
	return &CountHandler{service: svc}
}

type CountRequest struct {
	Polygon       []model.Coordinate `json:"polygon"`
	IncludedTypes []string           `json:"includedTypes"`
}

func (h *CountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	count, err := h.service.Count(req.Polygon, req.IncludedTypes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

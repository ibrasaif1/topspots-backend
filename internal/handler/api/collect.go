package api

import (
	"encoding/json"
	"net/http"

	"topspots-backend/internal/model"
	"topspots-backend/internal/service"
)

type CollectHandler struct {
	service *service.PlacesService
}

func NewCollectHandler(svc *service.PlacesService) *CollectHandler {
	return &CollectHandler{service: svc}
}

type CollectRequest struct {
	Polygon       []model.Coordinate `json:"polygon"`
	IncludedTypes []string           `json:"includedTypes"`
}

func (h *CollectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CollectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	places, err := h.service.CollectAndHydrate(r.Context(), req.Polygon, req.IncludedTypes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"places": places})
}

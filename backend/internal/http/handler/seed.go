package handler

import (
	"encoding/json"
	"net/http"
)

type SeedService interface {
	SeedDefaultAdmin() (email string, created bool, err error)
}

type SeedHandler struct {
	service SeedService
}

func NewSeedHandler(service SeedService) *SeedHandler {
	return &SeedHandler{service: service}
}

func (h *SeedHandler) SeedDefaultAdmin(w http.ResponseWriter, r *http.Request) {
	email, created, err := h.seed(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"email":   email,
			"created": created,
		},
		"meta":  map[string]any{},
		"error": nil,
	})
}

func (h *SeedHandler) seed(r *http.Request) (string, bool, error) {
	if h.service == nil {
		return "", false, nil
	}
	return h.service.SeedDefaultAdmin()
}

func init() {
	_ = json.Marshal
}

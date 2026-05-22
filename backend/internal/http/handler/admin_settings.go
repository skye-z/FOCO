package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"foco/backend/api/internal/domain/platform"
)

type AdminSettingsService interface {
	GetAdminSettings(ctx context.Context) (*platform.AdminSettings, error)
	UpdateAdminSettings(ctx context.Context, input platform.AdminSettingsUpdate) (*platform.AdminSettings, error)
	GetPublicSettings(ctx context.Context) (*platform.PublicSettings, error)
}

type AdminSettingsHandler struct {
	service AdminSettingsService
}

func NewAdminSettingsHandler(service AdminSettingsService) *AdminSettingsHandler {
	return &AdminSettingsHandler{service: service}
}

func (h *AdminSettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !requireAdminRole(w, r) {
		return
	}
	if h.service == nil {
		http.Error(w, "settings service unavailable", http.StatusInternalServerError)
		return
	}
	settings, err := h.service.GetAdminSettings(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": settings, "meta": map[string]any{}, "error": nil})
}

func (h *AdminSettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !requireAdminRole(w, r) {
		return
	}
	if h.service == nil {
		http.Error(w, "settings service unavailable", http.StatusInternalServerError)
		return
	}
	var body platform.AdminSettingsUpdate
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	settings, err := h.service.UpdateAdminSettings(r.Context(), body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": settings, "meta": map[string]any{}, "error": nil})
}

func (h *AdminSettingsHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		http.Error(w, "settings service unavailable", http.StatusInternalServerError)
		return
	}
	settings, err := h.service.GetPublicSettings(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": settings, "meta": map[string]any{}, "error": nil})
}

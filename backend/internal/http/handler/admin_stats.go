package handler

import (
	"context"
	"net/http"
	"time"

	"foco/backend/api/internal/domain/account"
)

const platformVersion = "0.2.0"

type StatsService interface {
	GetPlatformStats(ctx context.Context) (*account.PlatformStats, error)
}

type AdminStatsHandler struct {
	statsService StatsService
}

func NewAdminStatsHandler(statsService StatsService) *AdminStatsHandler {
	return &AdminStatsHandler{statsService: statsService}
}

func (h *AdminStatsHandler) Overview(w http.ResponseWriter, r *http.Request) {
	if h.statsService == nil {
		http.Error(w, "stats service unavailable", http.StatusInternalServerError)
		return
	}

	stats, err := h.statsService.GetPlatformStats(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"total_exams":      stats.TotalExams,
			"total_users":      stats.TotalUsers,
			"active_users_7d":  stats.ActiveUsers7Days,
			"platform_version": platformVersion,
			"last_updated":     time.Now().UTC().Format(time.RFC3339),
		},
		"meta":  map[string]any{},
		"error": nil,
	})
}

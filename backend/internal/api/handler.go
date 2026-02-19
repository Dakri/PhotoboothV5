package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"photobooth/internal/app"
	"photobooth/internal/logging"
)

type Handler struct {
	app *app.App
}

func NewHandler(a *app.App) *Handler {
	return &Handler{app: a}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/status", h.handleStatus)
	mux.HandleFunc("/api/trigger", h.handleTrigger)
	mux.HandleFunc("/api/photos", h.handlePhotos)
	mux.HandleFunc("/api/photos/latest", h.handleLatestPhoto)
	mux.HandleFunc("/api/logs", h.handleLogs)
	mux.HandleFunc("/api/legacy/poll", h.handleLegacyPoll)
}

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"state":   h.app.GetState(),
		"clients": h.app.Hub.ClientCount(),
		"uptime":  h.app.GetUptime(),
		"camera":  h.app.Camera.GetInfo(),
	}
	jsonResponse(w, status)
}

func (h *Handler) handleTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.app.Trigger()
	jsonResponse(w, map[string]string{"status": "triggered"})
}

func (h *Handler) handlePhotos(w http.ResponseWriter, r *http.Request) {
	photos, err := h.app.Storage.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, photos)
}

func (h *Handler) handleLatestPhoto(w http.ResponseWriter, r *http.Request) {
	photo := h.app.GetLastPhoto()
	if photo == nil {
		http.Error(w, "No photos found", http.StatusNotFound)
		return
	}
	jsonResponse(w, photo)
}

func (h *Handler) handleLogs(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			limit = n
		}
	}
	entries := logging.Get().GetEntries(limit)
	jsonResponse(w, entries)
}

func (h *Handler) handleLegacyPoll(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"state":     h.app.GetState(),
		"lastPhoto": h.app.GetLastPhoto(),
		"timestamp": time.Now().UnixMilli(),
	}
	jsonResponse(w, status)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

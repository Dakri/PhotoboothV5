package api

import (
	"encoding/json"
	"net/http"
	"time"

	"photobooth/internal/app"
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
	mux.HandleFunc("/api/legacy/poll", h.handleLegacyPoll)
}

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"state":   h.app.GetState(),
		"clients": len(h.app.Hub.Clients),
		"uptime":  "TODO",
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
	// Simple pagination could go here
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

func (h *Handler) handleLegacyPoll(w http.ResponseWriter, r *http.Request) {
	// Combined status for polling clients
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

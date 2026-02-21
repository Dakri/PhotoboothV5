package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"photobooth/internal/app"
	"photobooth/internal/config"
	"photobooth/internal/disk"
	"photobooth/internal/logging"
	"photobooth/internal/websocket"
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
	mux.HandleFunc("/api/settings", h.handleSettings)
	mux.HandleFunc("/api/legacy/poll", h.handleLegacyPoll)
	mux.HandleFunc("/api/gallery/count", h.handleGalleryCount)
	mux.HandleFunc("/api/gallery/empty", h.handleGalleryEmpty)
	mux.HandleFunc("/api/gallery/delete", h.handleGalleryDelete)
	mux.HandleFunc("/api/usb/devices", h.handleUsbDevices)
	mux.HandleFunc("/api/usb/export", h.handleUsbExport)
	mux.HandleFunc("/api/camera/files", h.handleCameraFiles)
}

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	// Disk Usage
	usage, err := disk.GetUsage(h.app.Config.Booth.PhotosBasePath)
	if err != nil {
		h.app.Log.Warn("system", "Failed to get disk usage: %v", err)
	}

	status := map[string]interface{}{
		"state":     h.app.GetState(),
		"clients":   h.app.Hub.ClientCount(),
		"uptime":    h.app.GetUptime(),
		"camera":    h.app.Camera.GetCachedInfo(),
		"disk":      usage,
		"lastPhoto": h.app.GetLastPhoto(),
	}
	jsonResponse(w, status)
}

func (h *Handler) handleCameraFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	files, err := h.app.Camera.ListCameraFiles()
	if err != nil {
		h.app.Log.Error("api", "Failed to list camera files: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, files)
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

func (h *Handler) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.getSettings(w, r)
	case "POST":
		h.postSettings(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, map[string]interface{}{
		"booth":  h.app.Config.Booth,
		"albums": h.app.ListAlbums(),
	})
}

func (h *Handler) postSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CountdownSeconds      *int    `json:"countdownSeconds"`
		PreviewDisplaySeconds *int    `json:"previewDisplaySeconds"`
		TriggerDelayMs        *int    `json:"triggerDelayMs"`
		CurrentAlbum          *string `json:"currentAlbum"`
		CaptureStrategy       *string `json:"captureStrategy"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	booth := h.app.Config.Booth

	if req.CountdownSeconds != nil {
		v := *req.CountdownSeconds
		if v < 1 {
			v = 1
		}
		if v > 10 {
			v = 10
		}
		booth.CountdownSeconds = v
	}

	if req.PreviewDisplaySeconds != nil {
		v := *req.PreviewDisplaySeconds
		if v < 1 {
			v = 1
		}
		if v > 30 {
			v = 30
		}
		booth.PreviewDisplaySeconds = v
	}

	if req.TriggerDelayMs != nil {
		v := *req.TriggerDelayMs
		if v < -3000 {
			v = -3000
		}
		if v > 1000 {
			v = 1000
		}
		booth.TriggerDelayMs = v
	}

	if req.CaptureStrategy != nil {
		v := strings.ToUpper(strings.TrimSpace(*req.CaptureStrategy))
		switch v {
		case "A", "B", "C", "D":
			if booth.AlbumCaptureMethods == nil {
				booth.AlbumCaptureMethods = make(map[string]string)
			}

			// Determine which album this strategy belongs to
			albumToUpdate := booth.CurrentAlbum
			if req.CurrentAlbum != nil && *req.CurrentAlbum != "" {
				albumToUpdate = config.SanitizeAlbumName(*req.CurrentAlbum)
			}
			booth.AlbumCaptureMethods[albumToUpdate] = v

			// Do not override global CaptureStrategy
			h.app.Camera.SetStrategy(v)
		default:
			http.Error(w, "captureStrategy must be A, B, C or D", http.StatusBadRequest)
			return
		}
	}

	// Apply struct changes first (countdown, preview, strategy)
	h.app.Config.UpdateBooth(booth)

	// Call SetAlbum after UpdateBooth so it doesn't get overwritten by the struct value copy
	if req.CurrentAlbum != nil && *req.CurrentAlbum != "" {
		h.app.SetAlbum(*req.CurrentAlbum)
	}

	// Save to disk
	if err := h.app.Config.Save(); err != nil {
		h.app.Log.Error("settings", "Failed to save config: %v", err)
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	h.app.Log.Info("settings", "Settings updated: countdown=%ds, preview=%ds, delay=%dms, album=%s",
		h.app.Config.Booth.CountdownSeconds, h.app.Config.Booth.PreviewDisplaySeconds, h.app.Config.Booth.TriggerDelayMs, h.app.Config.Booth.CurrentAlbum)

	jsonResponse(w, map[string]interface{}{
		"booth":  h.app.Config.Booth,
		"albums": h.app.ListAlbums(),
	})
}

func (h *Handler) handleLegacyPoll(w http.ResponseWriter, r *http.Request) {
	remaining, total := h.app.GetCountdown()
	status := map[string]interface{}{
		"state":     h.app.GetState(),
		"lastPhoto": h.app.GetLastPhoto(),
		"countdown": map[string]interface{}{"remaining": remaining, "total": total},
		"timestamp": time.Now().UnixMilli(),
	}
	jsonResponse(w, status)
}

func (h *Handler) handleGalleryCount(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("album")
	if name == "" {
		name = h.app.Config.Booth.CurrentAlbum
	}
	count, err := h.app.GetGalleryCount(name)
	if err != nil {
		h.app.Log.Error("api", "Failed to get gallery count for %s: %v", name, err)
		http.Error(w, "Failed to get count", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]int{"count": count})
}

func (h *Handler) handleGalleryEmpty(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("album")
	if name == "" {
		name = h.app.Config.Booth.CurrentAlbum
	}
	if err := h.app.EmptyGallery(name); err != nil {
		h.app.Log.Error("api", "Failed to empty gallery %s: %v", name, err)
		http.Error(w, "Failed to empty gallery", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"status": "emptied"})
}

func (h *Handler) handleGalleryDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("album")
	if name == "" {
		http.Error(w, "Album name required", http.StatusBadRequest)
		return
	}
	if err := h.app.DeleteGallery(name); err != nil {
		h.app.Log.Error("api", "Failed to delete gallery %s: %v", name, err)
		http.Error(w, err.Error(), http.StatusBadRequest) // e.g. "cannot delete default"
		return
	}
	jsonResponse(w, map[string]string{"status": "deleted"})
}

func (h *Handler) handleUsbDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := disk.GetUsbDevices()
	if err != nil {
		h.app.Log.Error("usb", "Failed to list USB devices: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, devices)
}

func (h *Handler) handleUsbExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DeviceName string `json:"deviceName"` // e.g., "sda1"
		AlbumName  string `json:"albumName"`
		CopyMode   string `json:"copyMode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.DeviceName == "" || req.AlbumName == "" {
		http.Error(w, "deviceName and albumName required", http.StatusBadRequest)
		return
	}

	sanitizedAlbum := config.SanitizeAlbumName(req.AlbumName)
	albumDir := filepath.Join(h.app.Config.Booth.PhotosBasePath, sanitizedAlbum)

	// Run export asynchronously so we don't block the request for huge albums
	go func() {
		h.app.Log.Info("usb", "Starting export of album '%s' to USB device '%s'...", req.AlbumName, req.DeviceName)

		h.app.Hub.Broadcast <- websocket.Event{
			Type:      "usb_export_start",
			Data:      map[string]string{"album": req.AlbumName},
			Timestamp: time.Now().UnixMilli(),
		}

		// 1. Mount device
		mountPoint, err := disk.MountUsb(req.DeviceName)
		if err != nil {
			h.app.Log.Error("usb", "Failed to mount device %s: %v", req.DeviceName, err)
			h.app.Hub.Broadcast <- websocket.Event{Type: "usb_export_error", Data: map[string]string{"message": "Mount failed"}}
			return
		}

		// 2. Prepare destination directory (e.g., /media/sda1/Photobooth_Export/Hochzeit)
		dstDir := filepath.Join(mountPoint, "Photobooth_Export", req.AlbumName)

		// 3. Check if we need to fetch RAWs from camera
		if req.CopyMode == "raw_jpeg" {
			method := "C" // default
			if m, ok := h.app.Config.Booth.AlbumCaptureMethods[sanitizedAlbum]; ok && m != "" {
				method = m
			}

			if method == "A" {
				// RAWs are only on the camera SD card, download them directly
				rawDstDir := filepath.Join(dstDir, "RAW_FROM_CAMERA")
				h.app.Log.Info("usb", "Copy mode is raw_jpeg and method is A. Fetching RAWs directly from camera to %s", rawDstDir)

				err = h.app.Camera.DownloadAllRawToPath(rawDstDir, func(copied, total int) {
					// Broadcast progress for RAWs
					if copied%2 == 0 || copied == total {
						h.app.Hub.Broadcast <- websocket.Event{
							Type: "usb_export_progress",
							Data: map[string]interface{}{
								"album":  req.AlbumName,
								"copied": copied,
								"total":  total,
								"phase":  fmt.Sprintf("Download RAW (%d/%d)", copied, total),
							},
							Timestamp: time.Now().UnixMilli(),
						}
					}
				})
				if err != nil {
					h.app.Log.Error("usb", "Failed to download RAWs from camera: %v", err)
				}
			}
		}

		// 4. Copy files recursively with progress reporting
		err = disk.CopyDirWithProgress(albumDir, dstDir, func(copied, total int) {
			// Throttle progress updates to avoid spamming the websocket
			if copied%5 == 0 || copied == total {
				h.app.Hub.Broadcast <- websocket.Event{
					Type: "usb_export_progress",
					Data: map[string]interface{}{
						"album":  req.AlbumName,
						"copied": copied,
						"total":  total,
					},
					Timestamp: time.Now().UnixMilli(),
				}
			}
		})

		if err != nil {
			h.app.Log.Error("usb", "Failed to copy files to USB: %v", err)
			h.app.Hub.Broadcast <- websocket.Event{Type: "usb_export_error", Data: map[string]string{"message": "Copy failed"}}
			return
		}

		// 4. Unmount before reporting success to ensure data is written cleanly
		err = disk.UnmountUsb(mountPoint)
		if err != nil {
			h.app.Log.Warn("usb", "Export succeeded, but unmount failed: %v", err)
			// Proceed to success anyway, as the copy phase worked.
		} else {
			h.app.Log.Info("usb", "USB device %s cleanly unmounted.", mountPoint)
		}

		h.app.Log.Info("usb", "Export of album '%s' to USB successful.", req.AlbumName)
		h.app.Hub.Broadcast <- websocket.Event{
			Type: "usb_export_success",
			Data: map[string]interface{}{
				"album": req.AlbumName,
				"path":  dstDir,
			},
			Timestamp: time.Now().UnixMilli(),
		}
	}()

	jsonResponse(w, map[string]string{"status": "export_started"})
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

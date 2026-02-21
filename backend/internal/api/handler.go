package api

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"photobooth/internal/app"
	"photobooth/internal/config"
	"photobooth/internal/disk"
	"photobooth/internal/logging"
	"photobooth/internal/websocket"
)

type Handler struct {
	app *app.App

	// USB export concurrency guard
	exportMu     sync.Mutex
	exportActive bool
	exportCancel context.CancelFunc
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
	mux.HandleFunc("/api/usb/export/cancel", h.handleUsbExportCancel)
	mux.HandleFunc("/api/usb/unmount", h.handleUsbUnmount)
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

	// Auto-mount any unmounted USB devices so we can read their free space
	for _, d := range devices {
		if d.MountPoint == "" {
			h.app.Log.Info("usb", "Auto-mounting %s...", d.Name)
			if _, err := disk.MountUsb(d.Name); err != nil {
				h.app.Log.Warn("usb", "Auto-mount of %s failed: %v", d.Name, err)
			}
		}
	}

	// Re-fetch after mounting so free space is populated
	devices, err = disk.GetUsbDevices()
	if err != nil {
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
		DeviceName string `json:"deviceName"`
		AlbumName  string `json:"albumName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.DeviceName == "" || req.AlbumName == "" {
		http.Error(w, "deviceName and albumName required", http.StatusBadRequest)
		return
	}

	// --- Concurrency guard: only one export at a time ---
	h.exportMu.Lock()
	if h.exportActive {
		h.exportMu.Unlock()
		http.Error(w, "An export is already running", http.StatusConflict)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	h.exportActive = true
	h.exportCancel = cancel
	h.exportMu.Unlock()

	sanitizedAlbum := config.SanitizeAlbumName(req.AlbumName)
	// Only copy the original folder
	srcDir := filepath.Join(h.app.Config.Booth.PhotosBasePath, sanitizedAlbum, "original")

	go func() {
		defer func() {
			h.exportMu.Lock()
			h.exportActive = false
			h.exportCancel = nil
			h.exportMu.Unlock()
			cancel() // always release resources
		}()

		h.app.Log.Info("usb", "Starting export of album '%s' (original only) to USB device '%s'...", req.AlbumName, req.DeviceName)
		h.app.Hub.Broadcast <- websocket.Event{
			Type:      "usb_export_start",
			Data:      map[string]string{"album": req.AlbumName},
			Timestamp: time.Now().UnixMilli(),
		}

		// 1. Mount device
		mountPoint, err := disk.MountUsb(req.DeviceName)
		if err != nil {
			h.app.Log.Error("usb", "Failed to mount device %s: %v", req.DeviceName, err)
			h.app.Hub.Broadcast <- websocket.Event{Type: "usb_export_error", Data: map[string]string{"message": "Mount failed: " + err.Error()}, Timestamp: time.Now().UnixMilli()}
			return
		}

		// 2. Destination
		dstDir := filepath.Join(mountPoint, "Photobooth_Export", req.AlbumName)

		// Track start time for ETA
		startTime := time.Now()

		// 3. Copy just the original folder with progress
		err = disk.CopyDirWithProgress(ctx, srcDir, dstDir, func(copiedBytes, totalBytes, copiedFiles, totalFiles int64) {
			var etaSecs int64
			if copiedBytes > 0 && totalBytes > 0 {
				elapsed := time.Since(startTime).Seconds()
				bytesPerSec := float64(copiedBytes) / elapsed
				if bytesPerSec > 0 {
					etaSecs = int64(float64(totalBytes-copiedBytes) / bytesPerSec)
				}
			}
			h.app.Hub.Broadcast <- websocket.Event{
				Type: "usb_export_progress",
				Data: map[string]interface{}{
					"album":       req.AlbumName,
					"copiedBytes": copiedBytes,
					"totalBytes":  totalBytes,
					"copiedFiles": copiedFiles,
					"totalFiles":  totalFiles,
					"etaSeconds":  etaSecs,
				},
				Timestamp: time.Now().UnixMilli(),
			}
		})

		if err != nil {
			msg := "Copy failed"
			if ctx.Err() != nil {
				msg = "Export cancelled"
				h.app.Log.Info("usb", "Export of album '%s' was cancelled.", req.AlbumName)
			} else {
				h.app.Log.Error("usb", "Failed to copy files to USB: %v", err)
			}
			h.app.Hub.Broadcast <- websocket.Event{Type: "usb_export_error", Data: map[string]string{"message": msg}, Timestamp: time.Now().UnixMilli()}
			return
		}

		// 4. Sync + do NOT auto-unmount – let the user press "Safely Remove"
		// Just flush buffers
		h.app.Log.Info("usb", "Export done. Flushing buffers...")
		h.app.Hub.Broadcast <- websocket.Event{
			Type:      "usb_export_success",
			Data:      map[string]interface{}{"album": req.AlbumName, "path": dstDir},
			Timestamp: time.Now().UnixMilli(),
		}
	}()

	jsonResponse(w, map[string]string{"status": "export_started"})
}

func (h *Handler) handleUsbExportCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.exportMu.Lock()
	active := h.exportActive
	cancel := h.exportCancel
	h.exportMu.Unlock()

	if !active || cancel == nil {
		http.Error(w, "No active export", http.StatusConflict)
		return
	}
	cancel()
	jsonResponse(w, map[string]string{"status": "cancelling"})
}

func (h *Handler) handleUsbUnmount(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DeviceName string `json:"deviceName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DeviceName == "" {
		http.Error(w, "deviceName required", http.StatusBadRequest)
		return
	}

	// Prevent unmounting a device that's being exported to
	h.exportMu.Lock()
	active := h.exportActive
	h.exportMu.Unlock()
	if active {
		http.Error(w, "Cannot unmount while export is running", http.StatusConflict)
		return
	}

	// Resolve mountpoint from device name
	devices, err := disk.GetUsbDevices()
	if err != nil {
		http.Error(w, "Failed to list devices: "+err.Error(), http.StatusInternalServerError)
		return
	}
	mountPoint := ""
	for _, d := range devices {
		if d.Name == req.DeviceName {
			mountPoint = d.MountPoint
			break
		}
	}
	if mountPoint == "" {
		http.Error(w, "Device not mounted or not found", http.StatusNotFound)
		return
	}

	if err := disk.UnmountUsb(mountPoint); err != nil {
		h.app.Log.Error("usb", "Unmount of %s failed: %v", req.DeviceName, err)
		http.Error(w, "Unmount failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	h.app.Log.Info("usb", "Device %s safely unmounted.", req.DeviceName)
	jsonResponse(w, map[string]string{"status": "unmounted"})
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"photobooth/internal/camera"
	"photobooth/internal/config"
	"photobooth/internal/disk"
	"photobooth/internal/imaging"
	"photobooth/internal/logging"
	"photobooth/internal/storage"
	"photobooth/internal/websocket"
)

type State string

const (
	StateIdle       State = "idle"
	StateCountdown  State = "countdown"
	StateCapturing  State = "capturing"
	StateProcessing State = "processing"
	StatePreview    State = "preview"
	StateError      State = "error"
)

type App struct {
	Config  *config.Config
	Camera  *camera.Controller
	Imaging *imaging.Processor
	Storage *storage.Manager
	Hub     *websocket.Hub
	Log     *logging.Logger

	mu                 sync.Mutex
	state              State
	lastPhoto          *storage.Photo
	startTime          time.Time
	countdownRemaining int
	countdownTotal     int
	captureSeq         int
}

func NewApp(cfg *config.Config, cam *camera.Controller, img *imaging.Processor, store *storage.Manager, hub *websocket.Hub) *App {
	logger := logging.Get()

	app := &App{
		Config:    cfg,
		Camera:    cam,
		Imaging:   img,
		Storage:   store,
		Hub:       hub,
		Log:       logger,
		state:     StateIdle,
		startTime: time.Now(),
	}

	// Wire up Hub events
	hub.OnTrigger = app.Trigger

	// Wire up logging broadcast via WebSocket
	logger.SetBroadcast(func(entry logging.Entry) {
		hub.Broadcast <- websocket.Event{
			Type:      websocket.EventTypeLog,
			Data:      entry,
			Timestamp: entry.Timestamp,
		}
	})

	logger.Info("system", "Photobooth application initialized")

	// Apply persisted capture strategy for the default/current album initially
	if s, ok := cfg.Booth.AlbumCaptureMethods[cfg.Booth.CurrentAlbum]; ok && s != "" {
		cam.SetStrategy(s)
	}

	// Start background camera info refresh (only when idle)
	go app.cameraInfoRefreshLoop()

	return app
}

func (a *App) GetState() State {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.state
}

func (a *App) GetUptime() string {
	d := time.Since(a.startTime)
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func (a *App) SetState(s State) {
	a.mu.Lock()
	a.state = s
	a.mu.Unlock()

	a.Hub.Broadcast <- websocket.Event{
		Type:      websocket.EventTypeStatus,
		Data:      map[string]interface{}{"state": s},
		Timestamp: time.Now().UnixMilli(),
	}
}

func (a *App) GetCountdown() (int, int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.countdownRemaining, a.countdownTotal
}

func (a *App) Trigger() {
	a.mu.Lock()
	if a.state != StateIdle && a.state != StatePreview {
		a.mu.Unlock()
		a.Log.Warn("trigger", "Trigger ignored: system not idle or preview (state: %s)", a.state)
		return
	}
	a.captureSeq++
	currentSeq := a.captureSeq
	a.mu.Unlock()

	a.SetState(StateCountdown)
	a.Log.Info("trigger", "Capture sequence %d started", currentSeq)

	go a.runCaptureSequence(currentSeq)
}

func (a *App) runCaptureSequence(seq int) {
	// 1. Countdown
	seconds := a.Config.Booth.CountdownSeconds
	if seconds < 1 {
		seconds = 3
	}

	activeAlbum := a.Config.Booth.CurrentAlbum
	method := "C" // Default strategy
	if m, ok := a.Config.Booth.AlbumCaptureMethods[activeAlbum]; ok && m != "" {
		method = m
	}
	a.Camera.SetStrategy(method)

	a.mu.Lock()
	a.countdownTotal = seconds
	a.mu.Unlock()

	a.Log.Info("countdown", "Countdown started: %d seconds", seconds)

	// Pre-configure camera (e.g. set capturetarget) while countdown is running
	a.Camera.PrepareCapture()

	var filename string
	var err error
	var captureDuration time.Duration

	delay := a.Config.Booth.TriggerDelayMs

	// Calculate absolute trigger time relative to the start of the countdown.
	// E.g., 3s countdown = 3000ms. Delay = -1500ms means trigger at 1500ms.
	totalDurationMs := seconds * 1000
	triggerOffsetMs := totalDurationMs + delay // delay can be negative or positive

	if triggerOffsetMs < 0 {
		triggerOffsetMs = 0 // Don't allow triggering before we even start
	}

	type capRes struct {
		f string
		e error
		d time.Duration
	}
	captureChan := make(chan capRes, 1)

	// Launch a background routine to fire the physical camera flash at the exact calculated offset
	go func() {
		if triggerOffsetMs > 0 {
			time.Sleep(time.Duration(triggerOffsetMs) * time.Millisecond)
		}
		t0 := time.Now()
		a.Log.Info("camera", "Physical trigger fired (Offset = %d ms, Delay = %d ms)...", triggerOffsetMs, delay)
		fname, cerr := a.Camera.Capture()
		captureChan <- capRes{fname, cerr, time.Since(t0)}
	}()

	// Run the VISUAL countdown completely undisturbed on a strictly 1000ms tick.
	for i := seconds; i > 0; i-- {
		a.mu.Lock()
		a.countdownRemaining = i
		a.mu.Unlock()

		a.Hub.Broadcast <- websocket.Event{
			Type:      websocket.EventTypeCountdown,
			Data:      map[string]interface{}{"remaining": i, "total": seconds},
			Timestamp: time.Now().UnixMilli(),
		}

		time.Sleep(1 * time.Second)
	}

	// The visual countdown has reached 0.
	a.mu.Lock()
	a.countdownRemaining = 0
	a.mu.Unlock()

	a.Hub.Broadcast <- websocket.Event{
		Type:      websocket.EventTypeCountdown,
		Data:      map[string]interface{}{"remaining": 0, "total": seconds},
		Timestamp: time.Now().UnixMilli(),
	}

	// Set state to "Bitte lächeln"
	a.SetState(StateCapturing)

	// Now wait for the background capture routine to finish taking the photo!
	res := <-captureChan
	filename = res.f
	err = res.e
	captureDuration = res.d

	if err != nil {
		a.Log.Error("camera", "Capture failed: %v", err)
		a.SetState(StateError)
		a.Hub.Broadcast <- websocket.Event{Type: "error", Data: map[string]string{"message": err.Error()}, Timestamp: time.Now().UnixMilli()}
		time.Sleep(2 * time.Second)
		a.SetState(StateIdle)
		return
	}

	a.Log.Info("camera", "Photo captured: %s", filename)

	// 3. Processing
	a.SetState(StateProcessing)
	a.Log.Info("imaging", "Processing image: %s", filename)

	albumDir := a.GetAlbumDir()
	fullPath := filepath.Join(albumDir, "original", filename)

	t1 := time.Now()
	var previewDuration time.Duration

	// Process image with callback for early preview
	err = a.Imaging.Process(fullPath, func() {
		// This callback runs as soon as the preview is ready (before thumbnail)
		previewDuration = time.Since(t1)

		a.lastPhoto = &storage.Photo{
			Filename:  filename,
			Url:       "/photos/preview/" + filename,
			ThumbUrl:  "/photos/thumb/" + filename,
			Timestamp: time.Now(),
		}

		// 4. Preview (Broadcast immediately)
		a.SetState(StatePreview)

		previewSecs := a.Config.Booth.PreviewDisplaySeconds
		if previewSecs < 1 {
			previewSecs = 5
		}

		a.Log.Info("preview", "Preview ready in %.3fs. Showing for %d seconds", previewDuration.Seconds(), previewSecs)
		a.Hub.Broadcast <- websocket.Event{
			Type:      websocket.EventTypePhoto,
			Data:      a.lastPhoto,
			Timestamp: time.Now().UnixMilli(),
		}

		// Verify if RAW/Backup exists on camera (Async)
		go func(fname string) {
			// Short delay to ensure camera is ready
			time.Sleep(500 * time.Millisecond)
			exists, err := a.Camera.VerifyLastCapture()
			if err == nil && !exists {
				a.Log.Error("camera", "CRITICAL: RAW/Backup for %s MISSING on camera!", fname)
			} else if exists && method == "B" {
				// Strategy B: Async RAW download
				a.Log.Info("camera", "Strategy B: Downloading RAW in background for %s", fname)
				err := a.Camera.DownloadLatestRaw(albumDir)
				if err != nil {
					a.Log.Error("camera", "Failed to download RAW: %v", err)
				}
			}
		}(filename)
	})

	processingDuration := time.Since(t1)

	if err != nil {
		a.Log.Error("imaging", "Processing failed: %v", err)
	} else {
		// Log detailed stats
		a.Log.Info("stats", "Capture=%.3fs, Preview=%.3fs, TotalProcessing=%.3fs",
			captureDuration.Seconds(), previewDuration.Seconds(), processingDuration.Seconds())
	}

	// Wait for preview duration (minus the time we already spent processing thumbnail)
	// We want to show the preview for at least 'PreviewDisplaySeconds' starting from when it appeared.
	// Since the callback runs in a goroutine, we need to handle the sleep here carefully.
	// The simplest approach for now is to just sleep the full duration from this point,
	// effectively making the total time "Processing Time + Preview Time".

	previewSecs := a.Config.Booth.PreviewDisplaySeconds
	if previewSecs < 1 {
		previewSecs = 5
	}
	time.Sleep(time.Duration(previewSecs) * time.Second)

	// 5. Finish
	a.mu.Lock()
	if a.state == StatePreview && a.captureSeq == seq {
		a.state = StateIdle
		a.mu.Unlock()

		a.Hub.Broadcast <- websocket.Event{
			Type:      websocket.EventTypeStatus,
			Data:      map[string]interface{}{"state": StateIdle},
			Timestamp: time.Now().UnixMilli(),
		}
		a.Log.Info("system", "Capture sequence complete. Ready for next trigger.")
	} else {
		// State was changed by another trigger (e.g. preview skipped)
		a.mu.Unlock()
	}
}

func (a *App) GetLastPhoto() *storage.Photo {
	if a.lastPhoto != nil {
		return a.lastPhoto
	}
	return a.Storage.GetLatest()
}

// cameraInfoRefreshLoop refreshes camera info dynamically.
// If camera is connected: every 10s.
// If camera is disconnected: every 2s (to detect it faster).
func (a *App) cameraInfoRefreshLoop() {
	// Initial refresh
	time.Sleep(2 * time.Second)
	a.Camera.RefreshInfo()

	for {
		interval := 10 * time.Second
		if !a.Camera.IsConnected() {
			interval = 2 * time.Second
		}

		time.Sleep(interval)

		if a.GetState() == StateIdle {
			a.Camera.RefreshInfo()
		}

		// Broadcast System Info (Camera + Disk)
		usage, _ := disk.GetUsage(a.Config.Booth.PhotosBasePath)
		a.Hub.Broadcast <- websocket.Event{
			Type: websocket.EventTypeSystem,
			Data: map[string]interface{}{
				"camera": a.Camera.GetCachedInfo(),
				"disk":   usage,
			},
			Timestamp: time.Now().UnixMilli(),
		}
	}
}

// GetAlbumDir returns the full path to the current album directory.
func (a *App) GetAlbumDir() string {
	album := config.SanitizeAlbumName(a.Config.Booth.CurrentAlbum)
	return filepath.Join(a.Config.Booth.PhotosBasePath, album)
}

// EnsureAlbumDirs creates the original/preview/thumb subdirs for the current album.
func (a *App) EnsureAlbumDirs() {
	base := a.GetAlbumDir()
	for _, sub := range []string{"original", "preview", "thumb"} {
		os.MkdirAll(filepath.Join(base, sub), 0755)
	}
}

type AlbumInfo struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Count         int    `json:"count"`
	Size          int64  `json:"size"`
	CaptureMethod string `json:"captureMethod"`
}

// ListAlbums returns all existing albums with their original display name.
func (a *App) ListAlbums() []AlbumInfo {
	entries, err := os.ReadDir(a.Config.Booth.PhotosBasePath)
	if err != nil {
		return []AlbumInfo{}
	}

	var albums []AlbumInfo

	for _, e := range entries {
		if e.IsDir() {
			sanitized := e.Name()
			// Filter out legacy/system folders
			if sanitized == "original" || sanitized == "preview" || sanitized == "thumb" || sanitized == "css" || sanitized == "js" {
				continue
			}

			// Get original name from config map, fallback to sanitized name
			originalName := sanitized
			if name, ok := a.Config.Booth.AlbumDisplayNames[sanitized]; ok {
				originalName = name
			}

			// Get capture method
			captureMethod := "C" // Default strategy
			if method, ok := a.Config.Booth.AlbumCaptureMethods[sanitized]; ok {
				captureMethod = method
			}

			// optionally get count and size
			count, _ := a.GetGalleryCount(sanitized)
			size, _ := a.GetGallerySize(sanitized)

			albums = append(albums, AlbumInfo{
				Id:            sanitized,
				Name:          originalName,
				Count:         count,
				Size:          size,
				CaptureMethod: captureMethod,
			})
		}
	}
	return albums
}

// GetGallerySize returns the total size in bytes of the album's files.
func (a *App) GetGallerySize(name string) (int64, error) {
	sanitized := config.SanitizeAlbumName(name)
	base := filepath.Join(a.Config.Booth.PhotosBasePath, sanitized)

	var totalSize int64
	for _, sub := range []string{"original", "preview", "thumb"} {
		dir := filepath.Join(base, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				if info, err := e.Info(); err == nil {
					totalSize += info.Size()
				}
			}
		}
	}
	return totalSize, nil
}

// SetAlbum sets the current album, saves original name, creates directories, returns the sanitized name.
func (a *App) SetAlbum(originalName string) (string, error) {
	sanitized := config.SanitizeAlbumName(originalName)

	// Preserve existing display name if only the sanitized ID is passed
	if sanitized == originalName {
		if a.Config.Booth.AlbumDisplayNames != nil {
			if existing, ok := a.Config.Booth.AlbumDisplayNames[sanitized]; ok && existing != "" {
				originalName = existing
			}
		}
	}

	a.Config.Booth.CurrentAlbum = sanitized

	// Save original name
	if a.Config.Booth.AlbumDisplayNames == nil {
		a.Config.Booth.AlbumDisplayNames = make(map[string]string)
	}
	a.Config.Booth.AlbumDisplayNames[sanitized] = originalName

	// Update Camera data dir to point to the album
	albumDir := a.GetAlbumDir()
	a.Camera.SetDataDir(albumDir)
	a.Storage.SetRootDir(albumDir)
	a.EnsureAlbumDirs()

	a.Log.Info("settings", "Album set to '%s' (original: '%s', path: %s)", sanitized, originalName, albumDir)
	return sanitized, nil
}

// GetGalleryCount returns the number of images in the album's 'original' folder.
func (a *App) GetGalleryCount(name string) (int, error) {
	sanitized := config.SanitizeAlbumName(name)
	dir := filepath.Join(a.Config.Booth.PhotosBasePath, sanitized, "original")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if ext == ".jpg" || ext == ".jpeg" {
				count++
			}
		}
	}
	return count, nil
}

// EmptyGallery deletes all photos in the album but keeps the album itself.
func (a *App) EmptyGallery(name string) error {
	sanitized := config.SanitizeAlbumName(name)
	base := filepath.Join(a.Config.Booth.PhotosBasePath, sanitized)

	// Clean subdirs
	for _, sub := range []string{"original", "preview", "thumb"} {
		dir := filepath.Join(base, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			os.Remove(filepath.Join(dir, e.Name()))
		}
	}
	a.Log.Info("system", "Emptied gallery: %s", sanitized)
	return nil
}

// DeleteGallery removes the entire album directory and its entry in the settings.
func (a *App) DeleteGallery(name string) error {
	sanitized := config.SanitizeAlbumName(name)
	if sanitized == "default" {
		return fmt.Errorf("cannot delete default album")
	}
	if sanitized == a.Config.Booth.CurrentAlbum {
		return fmt.Errorf("cannot delete active album")
	}

	path := filepath.Join(a.Config.Booth.PhotosBasePath, sanitized)
	err := os.RemoveAll(path)
	if err == nil {
		if a.Config.Booth.AlbumDisplayNames != nil {
			delete(a.Config.Booth.AlbumDisplayNames, sanitized)
			a.Config.Save() // Save to persist the deletion from map
		}
		a.Log.Info("system", "Deleted gallery: %s", sanitized)
	}
	return err
}

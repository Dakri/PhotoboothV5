package app

import (
	"fmt"
	"sync"
	"time"

	"photobooth/internal/camera"
	"photobooth/internal/config"
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

	mu        sync.Mutex
	state     State
	lastPhoto *storage.Photo
	startTime time.Time
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

func (a *App) Trigger() {
	a.mu.Lock()
	if a.state != StateIdle {
		a.mu.Unlock()
		a.Log.Warn("trigger", "Trigger ignored: system not idle (state: %s)", a.state)
		return
	}
	a.state = StateCountdown
	a.mu.Unlock()

	a.Log.Info("trigger", "Capture sequence started")

	go a.runCaptureSequence()
}

func (a *App) runCaptureSequence() {
	// 1. Countdown
	seconds := 5
	a.Log.Info("countdown", "Countdown started: %d seconds", seconds)
	for i := seconds; i > 0; i-- {
		a.Hub.Broadcast <- websocket.Event{
			Type:      websocket.EventTypeCountdown,
			Data:      map[string]interface{}{"remaining": i, "total": seconds},
			Timestamp: time.Now().UnixMilli(),
		}
		time.Sleep(1 * time.Second)
	}

	a.Hub.Broadcast <- websocket.Event{
		Type:      websocket.EventTypeCountdown,
		Data:      map[string]interface{}{"remaining": 0, "total": seconds},
		Timestamp: time.Now().UnixMilli(),
	}

	// 2. Capture
	a.SetState(StateCapturing)
	a.Log.Info("camera", "Capturing photo...")

	filename, err := a.Camera.Capture()
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

	fullPath := "data/photos/original/" + filename

	if err := a.Imaging.Process(fullPath); err != nil {
		a.Log.Error("imaging", "Processing failed: %v", err)
	} else {
		a.Log.Info("imaging", "Image processed successfully")
	}

	a.lastPhoto = &storage.Photo{
		Filename:  filename,
		Url:       "/photos/preview/" + filename,
		ThumbUrl:  "/photos/thumb/" + filename,
		Timestamp: time.Now(),
	}

	// 4. Preview
	a.SetState(StatePreview)
	a.Log.Info("preview", "Showing preview for 8 seconds")
	a.Hub.Broadcast <- websocket.Event{
		Type:      websocket.EventTypePhoto,
		Data:      a.lastPhoto,
		Timestamp: time.Now().UnixMilli(),
	}

	time.Sleep(8 * time.Second)

	// 5. Finish
	a.SetState(StateIdle)
	a.Log.Info("system", "Capture sequence complete. Ready for next trigger.")
}

func (a *App) GetLastPhoto() *storage.Photo {
	if a.lastPhoto != nil {
		return a.lastPhoto
	}
	return a.Storage.GetLatest()
}

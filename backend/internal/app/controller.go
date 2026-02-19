package app

import (
	"log"
	"sync"
	"time"

	"photobooth/internal/camera"
	"photobooth/internal/config"
	"photobooth/internal/imaging"
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

	mu        sync.Mutex
	state     State
	lastPhoto *storage.Photo
}

func NewApp(cfg *config.Config, cam *camera.Controller, img *imaging.Processor, store *storage.Manager, hub *websocket.Hub) *App {
	app := &App{
		Config:  cfg,
		Camera:  cam,
		Imaging: img,
		Storage: store,
		Hub:     hub,
		state:   StateIdle,
	}

	// Wire up Hub events
	hub.OnTrigger = app.Trigger

	return app
}

func (a *App) GetState() State {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.state
}

func (a *App) SetState(s State) {
	a.mu.Lock()
	a.state = s
	a.mu.Unlock()

	// Broadcast state change
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
		// Send error to client that triggered? Or broadcast?
		// simple ignore for now or log
		log.Println("⚠️ Trigger ignored: not idle")
		return
	}
	a.state = StateCountdown
	a.mu.Unlock()

	log.Println("🚀 Capture sequence started")

	go a.runCaptureSequence()
}

func (a *App) runCaptureSequence() {
	// 1. Countdown
	seconds := 5 // TODO: Config
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
	a.Hub.Broadcast <- websocket.Event{Type: "capturing", Timestamp: time.Now().UnixMilli()}

	filename, err := a.Camera.Capture()
	if err != nil {
		log.Printf("❌ Capture error: %v", err)
		a.SetState(StateError)
		a.Hub.Broadcast <- websocket.Event{Type: "error", Data: map[string]string{"message": err.Error()}, Timestamp: time.Now().UnixMilli()}
		time.Sleep(2 * time.Second)
		a.SetState(StateIdle)
		return
	}

	// 3. Processing
	a.SetState(StateProcessing)
	a.Hub.Broadcast <- websocket.Event{Type: "processing", Timestamp: time.Now().UnixMilli()}

	// Assuming filename is just the name, construct full path for processing
	// Helper needed in storage or camera to get full path, or just hardcode for now based on knowledge
	// Camera returns "IMG_xxx.jpg", file is in data/photos/original/IMG_xxx.jpg
	fullPath := "data/photos/original/" + filename // TODO: Use config/helpers

	if err := a.Imaging.Process(fullPath); err != nil {
		log.Printf("❌ Processing error: %v", err)
		// Non-fatal? We have the original. But preview won't work.
	}

	// Update Photo List cache or similar?
	// refresh latest photo
	a.lastPhoto = &storage.Photo{
		Filename:  filename,
		Url:       "/photos/preview/" + filename,
		ThumbUrl:  "/photos/thumb/" + filename,
		Timestamp: time.Now(),
	}

	// 4. Preview
	a.SetState(StatePreview)
	a.Hub.Broadcast <- websocket.Event{
		Type:      websocket.EventTypePhoto,
		Data:      a.lastPhoto,
		Timestamp: time.Now().UnixMilli(),
	}

	// Wait for preview time
	time.Sleep(8 * time.Second) // TODO: Config

	// 5. Finish
	a.SetState(StateIdle)
}

func (a *App) GetLastPhoto() *storage.Photo {
	if a.lastPhoto != nil {
		return a.lastPhoto
	}
	return a.Storage.GetLatest()
}

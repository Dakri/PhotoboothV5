package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"photobooth/internal/api"
	"photobooth/internal/app"
	"photobooth/internal/camera"
	"photobooth/internal/config"
	"photobooth/internal/dns"
	"photobooth/internal/imaging"
	"photobooth/internal/logging"
	"photobooth/internal/network"
	"photobooth/internal/storage"
	"photobooth/internal/websocket"
)

func main() {
	// 1. Setup Logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("╔══════════════════════════════════════╗")
	log.Println("║       📷 PhotoboothV5 Starting       ║")
	log.Println("╚══════════════════════════════════════╝")

	// Initialize structured logger
	appLog := logging.Init(500)

	// 2. Load Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	appLog.Info("config", "Configuration loaded successfully")

	// 3. Prepare Directories
	cwd, _ := os.Getwd()
	dataDir := filepath.Join(cwd, "data", "photos")

	// 4. Initialize Components

	// Storage
	store := storage.NewManager(dataDir)
	store.EnsureDirs()

	// Camera
	cam := camera.NewController(cfg.Camera, dataDir)

	// Imaging
	img := imaging.NewProcessor(cfg.Image)

	// WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// App Controller (Orchestrator)
	application := app.NewApp(cfg, cam, img, store, hub)

	// Network (WiFi + DNS)
	if cfg.Wifi.Enabled {
		network.SetupWifi(cfg.Wifi)
		defer network.TeardownWifi() // Clean up on exit

		// Start Captive Portal DNS
		dnsServer := dns.NewServer(cfg.Wifi)
		dnsServer.Start()
		defer dnsServer.Stop()
	}

	// 5. Setup Routes
	mux := http.NewServeMux()

	// API
	apiHandler := api.NewHandler(application)
	// Manually register routes since I changed the signature in api/handler.go
	// Wait, I made api.Handler struct with RegisterRoutes method
	apiHandler.RegisterRoutes(mux)

	// WebSocket
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWs(w, r)
	})

	// Static Files (Frontend) - Served at root
	fs := http.FileServer(http.Dir("./public/frontend"))
	mux.Handle("/", fs)

	// Legacy Client - Served at /legacy/
	legacyFs := http.FileServer(http.Dir("./public/legacy"))
	mux.Handle("/legacy/", http.StripPrefix("/legacy/", legacyFs))

	// Photos - Served at /photos/
	photoFs := http.FileServer(http.Dir(dataDir))
	mux.Handle("/photos/", http.StripPrefix("/photos/", photoFs))

	srv := &http.Server{
		Addr:    ":80",
		Handler: mux,
	}

	// 6. Start Server
	go func() {
		log.Printf("🚀 Server running at http://localhost:80")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	// 7. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("🛑 Shutting down...")

	// WiFi Teardown & DNS Stop happens via defers here
	// Wait a bit for them to finish
	time.Sleep(1 * time.Second)
	log.Println("Server closed")
}

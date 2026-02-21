package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type Config struct {
	Wifi   WifiConfig   `json:"wifi"`
	Camera CameraConfig `json:"camera"`
	Image  ImageConfig  `json:"image"`
	Booth  BoothConfig  `json:"booth"`

	mu       sync.Mutex `json:"-"`
	filePath string     `json:"-"`
}

type WifiConfig struct {
	Enabled        bool   `json:"enabled"`
	Ssid           string `json:"ssid"`
	Password       string `json:"password"`
	Interface      string `json:"interface"`
	IpAddress      string `json:"ipAddress"`
	DhcpRangeStart string `json:"dhcpRangeStart"`
	DhcpRangeEnd   string `json:"dhcpRangeEnd"`
	CaptivePortal  bool   `json:"captivePortal"`
}

type CameraConfig struct {
	Enabled bool `json:"enabled"`
	Mock    bool `json:"mock"`
}

type ImageConfig struct {
	PreviewWidth   int  `json:"previewWidth"`
	PreviewQuality int  `json:"previewQuality"`
	ThumbWidth     int  `json:"thumbnailWidth"`
	ThumbQuality   int  `json:"thumbnailQuality"`
	KeepOriginal   bool `json:"keepOriginal"`
}

type BoothConfig struct {
	CountdownSeconds      int               `json:"countdownSeconds"`
	PreviewDisplaySeconds int               `json:"previewDisplaySeconds"`
	TriggerDelayMs        int               `json:"triggerDelayMs"` // Offset in ms. Negative triggers early.
	PhotosBasePath        string            `json:"photosBasePath"`
	CurrentAlbum          string            `json:"currentAlbum"`
	AlbumDisplayNames     map[string]string `json:"albumDisplayNames"`   // sanitized -> original
	AlbumCaptureMethods   map[string]string `json:"albumCaptureMethods"` // sanitized -> strategy (A, B, C)
}

func Load() (*Config, error) {
	// Default base values in case no file exists
	cfg := &Config{
		Wifi: WifiConfig{
			Enabled:        true,
			Ssid:           "Photobooth",
			Interface:      "wlan0",
			IpAddress:      "192.168.4.1",
			DhcpRangeStart: "192.168.4.10",
			DhcpRangeEnd:   "192.168.4.100",
			CaptivePortal:  true,
		},
		Camera: CameraConfig{
			Enabled: true,
			Mock:    false,
		},
		Image: ImageConfig{
			PreviewWidth:   1024,
			PreviewQuality: 70,
			ThumbWidth:     256,
			ThumbQuality:   70,
			KeepOriginal:   true,
		},
		Booth: BoothConfig{
			CountdownSeconds:      3,
			PreviewDisplaySeconds: 5,
			TriggerDelayMs:        0,
			PhotosBasePath:        "data/photos",
			CurrentAlbum:          "default",
			AlbumDisplayNames:     make(map[string]string),
			AlbumCaptureMethods:   make(map[string]string),
		},
	}
	cfg.Booth.AlbumDisplayNames["default"] = "Default"
	cfg.Booth.AlbumCaptureMethods["default"] = "C"

	// Paths to try
	serverConfPaths := []string{"../config/server.conf.json", "config/server.conf.json", "server.conf.json"}
	userConfPaths := []string{"../config/user.conf.json", "config/user.conf.json", "user.conf.json"}

	// 1. Load Server Config
	var activeConfigDir string
	for _, p := range serverConfPaths {
		if file, err := os.Open(p); err == nil {
			defer file.Close()
			json.NewDecoder(file).Decode(cfg)
			activeConfigDir = filepath.Dir(p)
			break
		}
	}

	// 2. Load User Config (Overrides server config)
	loadedUser := false
	for _, p := range userConfPaths {
		if file, err := os.Open(p); err == nil {
			defer file.Close()
			json.NewDecoder(file).Decode(cfg)
			cfg.filePath = p
			loadedUser = true
			break
		}
	}

	if !loadedUser {
		// Target the same directory where we found the server config, or default to local folder
		if activeConfigDir != "" {
			cfg.filePath = filepath.Join(activeConfigDir, "user.conf.json")
		} else {
			cfg.filePath = "user.conf.json"
		}
	}

	// Ensure maps exist if they were wiped by a bad JSON
	if cfg.Booth.AlbumDisplayNames == nil {
		cfg.Booth.AlbumDisplayNames = make(map[string]string)
		cfg.Booth.AlbumDisplayNames["default"] = "Default"
	}
	if cfg.Booth.AlbumCaptureMethods == nil {
		cfg.Booth.AlbumCaptureMethods = make(map[string]string)
	}
	if _, ok := cfg.Booth.AlbumCaptureMethods["default"]; !ok {
		cfg.Booth.AlbumCaptureMethods["default"] = "C"
	}

	return cfg, nil
}

// Save writes the current config to user.conf.json.
func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// Ensure the parent directory exists
	if dir := filepath.Dir(c.filePath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(c.filePath, data, 0644)
}

// UpdateBooth updates the booth config and saves.
func (c *Config) UpdateBooth(booth BoothConfig) {
	c.mu.Lock()
	c.Booth = booth
	c.mu.Unlock()
}

// SanitizeAlbumName converts a human-friendly album name to a filesystem-safe one.
// "Hoch Zeit!" → "hoch_zeit", "test  event" → "test_event"
func SanitizeAlbumName(name string) string {
	// Lowercase
	name = strings.ToLower(strings.TrimSpace(name))

	// Replace everything that's not a-z, 0-9, or - with underscore
	re := regexp.MustCompile(`[^a-z0-9\-]+`)
	name = re.ReplaceAllString(name, "_")

	// Remove leading/trailing underscores
	name = strings.Trim(name, "_")

	// Collapse multiple underscores
	re2 := regexp.MustCompile(`_+`)
	name = re2.ReplaceAllString(name, "_")

	if name == "" {
		name = "default"
	}
	return name
}

package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Wifi   WifiConfig   `json:"wifi"`
	Camera CameraConfig `json:"camera"`
	Image  ImageConfig  `json:"image"`
}

type WifiConfig struct {
	Enabled   bool   `json:"enabled"`
	Ssid      string `json:"ssid"`
	Password  string `json:"password"`
	Interface string `json:"interface"`
	IpAddress string `json:"ipAddress"`
}

type CameraConfig struct {
	Enabled bool `json:"enabled"`
	Mock    bool `json:"mock"`
}

type ImageConfig struct {
	PreviewWidth int `json:"previewWidth"`
	ThumbWidth   int `json:"thumbWidth"`
}

func Load() (*Config, error) {
	// Default values
	cfg := &Config{
		Wifi: WifiConfig{
			Enabled:   true,
			Ssid:      "Photobooth",
			Interface: "wlan0",
			IpAddress: "192.168.4.1",
		},
		Camera: CameraConfig{
			Enabled: true,
			Mock:    false,
		},
		Image: ImageConfig{
			PreviewWidth: 800,
			ThumbWidth:   200,
		},
	}

	// Try load from config.json
	file, err := os.Open("config.json")
	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

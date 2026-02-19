package network

import (
	"fmt"
	"log"
	"os/exec"
	"photobooth/internal/config"
)

func SetupWifi(cfg config.WifiConfig) {
	if !cfg.Enabled {
		return
	}

	log.Printf("📡 Setting up WiFi Hotspot: %s", cfg.Ssid)

	// Check for nmcli
	if _, err := exec.LookPath("nmcli"); err != nil {
		log.Println("⚠️ nmcli not found. WiFi setup skipped.")
		return
	}

	// Prepare cleanup just in case
	TeardownWifi()

	// Create connection
	// nmcli connection add type wifi ifname wlan0 con-name photobooth-ap autoconnect yes ssid "Photobooth"
	cmdAdd := exec.Command("nmcli", "connection", "add",
		"type", "wifi",
		"ifname", cfg.Interface,
		"con-name", "photobooth-ap",
		"autoconnect", "yes",
		"ssid", cfg.Ssid,
	)
	if out, err := cmdAdd.CombinedOutput(); err != nil {
		log.Printf("❌ Failed to add WiFi connection: %v\n%s", err, out)
		return
	}

	// Configure as AP
	// nmcli connection modify photobooth-ap 802-11-wireless.mode ap 802-11-wireless.band bg ipv4.addresses 192.168.4.1/24 ipv4.method shared
	cmdMod := exec.Command("nmcli", "connection", "modify", "photobooth-ap",
		"802-11-wireless.mode", "ap",
		"802-11-wireless.band", "bg",
		"ipv4.addresses", fmt.Sprintf("%s/24", cfg.IpAddress),
		"ipv4.method", "shared",
	)
	if out, err := cmdMod.CombinedOutput(); err != nil {
		log.Printf("❌ Failed to configure AP mode: %v\n%s", err, out)
		return
	}

	// Security (Open or WPA)
	if cfg.Password == "" {
		// Open network
		exec.Command("nmcli", "connection", "modify", "photobooth-ap", "remove", "802-11-wireless-security").Run()
	} else {
		// WPA
		exec.Command("nmcli", "connection", "modify", "photobooth-ap",
			"802-11-wireless-security.key-mgmt", "wpa-psk",
			"802-11-wireless-security.psk", cfg.Password).Run()
	}

	// Up
	if out, err := exec.Command("nmcli", "connection", "up", "photobooth-ap").CombinedOutput(); err != nil {
		log.Printf("❌ Failed to bring up WiFi: %v\n%s", err, out)
	} else {
		log.Printf("✅ WiFi Hotspot active: %s @ %s", cfg.Ssid, cfg.IpAddress)
	}
}

func TeardownWifi() {
	// nmcli connection delete photobooth-ap
	exec.Command("nmcli", "connection", "delete", "photobooth-ap").Run()
	log.Println("📡 WiFi Hotspot removed")
}

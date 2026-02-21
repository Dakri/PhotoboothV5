package network

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"photobooth/internal/config"
)

const connectionName = "photobooth-ap"

// connectionExists checks if the NetworkManager connection already exists.
func connectionExists(name string) bool {
	out, err := exec.Command("nmcli", "-t", "-f", "NAME", "connection", "show").CombinedOutput()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == name {
			return true
		}
	}
	return false
}

// connectionIsActive checks if the connection is currently active.
func connectionIsActive(name string) bool {
	out, err := exec.Command("nmcli", "-t", "-f", "NAME", "connection", "show", "--active").CombinedOutput()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == name {
			return true
		}
	}
	return false
}

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

	if connectionExists(connectionName) {
		log.Printf("   Connection '%s' already exists, updating settings...", connectionName)
	} else {
		log.Printf("   Creating new connection '%s'...", connectionName)
		// Create connection
		cmdAdd := exec.Command("nmcli", "connection", "add",
			"type", "wifi",
			"ifname", cfg.Interface,
			"con-name", connectionName,
			"autoconnect", "yes",
			"ssid", cfg.Ssid,
		)
		if out, err := cmdAdd.CombinedOutput(); err != nil {
			log.Printf("❌ Failed to add WiFi connection: %v\n%s", err, out)
			return
		}
	}

	// Configure / update as AP with DHCP
	modifyArgs := []string{"connection", "modify", connectionName,
		"802-11-wireless.mode", "ap",
		"802-11-wireless.band", "bg",
		"802-11-wireless.ssid", cfg.Ssid,
		"ipv4.addresses", fmt.Sprintf("%s/24", cfg.IpAddress),
		"ipv4.method", "shared",
		"connection.autoconnect", "yes",
	}

	if out, err := exec.Command("nmcli", modifyArgs...).CombinedOutput(); err != nil {
		log.Printf("❌ Failed to configure AP mode: %v\n%s", err, out)
		return
	}

	// Security (Open or WPA)
	if cfg.Password == "" {
		// Open network
		exec.Command("nmcli", "connection", "modify", connectionName, "remove", "802-11-wireless-security").Run()
	} else {
		// WPA
		exec.Command("nmcli", "connection", "modify", connectionName,
			"802-11-wireless-security.key-mgmt", "wpa-psk",
			"802-11-wireless-security.psk", cfg.Password).Run()
	}

	// Activate if not already active
	if connectionIsActive(connectionName) {
		log.Printf("   Connection '%s' is already active, reapplying...", connectionName)
		// Reapply to pick up any config changes
		if out, err := exec.Command("nmcli", "connection", "up", connectionName).CombinedOutput(); err != nil {
			log.Printf("⚠️ Failed to reapply WiFi: %v\n%s", err, out)
		}
	} else {
		if out, err := exec.Command("nmcli", "connection", "up", connectionName).CombinedOutput(); err != nil {
			log.Printf("❌ Failed to bring up WiFi: %v\n%s", err, out)
		} else {
			log.Printf("✅ WiFi Hotspot active: %s @ %s", cfg.Ssid, cfg.IpAddress)
		}
	}
}

func TeardownWifi() {
	// nmcli connection delete photobooth-ap
	exec.Command("nmcli", "connection", "delete", connectionName).Run()
	log.Println("📡 WiFi Hotspot removed")
}

package camera

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"photobooth/internal/config"
	"strings"
	"sync"
	"time"
)

// CameraInfo holds information about the connected camera.
type CameraInfo struct {
	Connected    bool   `json:"connected"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	SerialNumber string `json:"serialNumber"`
	LensName     string `json:"lensName"`
	BatteryLevel string `json:"batteryLevel"`
	StorageTotal string `json:"storageTotal"`
	StorageFree  string `json:"storageFree"`
}

// GetInfo queries the connected camera for summary and storage information.
// In mock mode it returns static dummy values.
func (c *Controller) GetInfo() CameraInfo {
	if c.config.Mock {
		return CameraInfo{
			Connected:    true,
			Model:        "Canon EOS 700D (Mock)",
			Manufacturer: "Canon Inc.",
			SerialNumber: "MOCK-123456",
			LensName:     "EF-S 18-55mm f/3.5-5.6 IS STM",
			BatteryLevel: "75%",
			StorageTotal: "32 GB",
			StorageFree:  "28 GB",
		}
	}

	info := CameraInfo{}

	// --- gphoto2 --summary ---
	summaryOut, err := exec.Command("gphoto2", "--summary").CombinedOutput()
	if err != nil {
		log.Printf("⚠️  Camera info unavailable: %v", err)
		return info // connected stays false
	}
	info.Connected = true
	parseSummary(string(summaryOut), &info)

	// --- gphoto2 --storage-info ---
	storageOut, err := exec.Command("gphoto2", "--storage-info").CombinedOutput()
	if err == nil {
		parseStorage(string(storageOut), &info)
	}

	return info
}

// parseSummary extracts key-value pairs from gphoto2 --summary output.
func parseSummary(output string, info *CameraInfo) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch strings.ToLower(key) {
		case "model":
			info.Model = val
		case "manufacturer":
			info.Manufacturer = val
		case "serial number":
			info.SerialNumber = val
		case "lens name":
			info.LensName = val
		case "battery level":
			info.BatteryLevel = val
		}
	}
}

// parseStorage extracts storage capacity from gphoto2 --storage-info output.
func parseStorage(output string, info *CameraInfo) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch strings.ToLower(key) {
		case "totalcapacity":
			info.StorageTotal = val
		case "free":
			info.StorageFree = val
		}
	}
}

type Controller struct {
	mu      sync.Mutex
	busy    bool
	config  config.CameraConfig
	dataDir string
}

func NewController(cfg config.CameraConfig, dataDir string) *Controller {
	return &Controller{
		config:  cfg,
		dataDir: dataDir,
	}
}

func (c *Controller) Capture() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.busy {
		return "", fmt.Errorf("camera is busy")
	}
	c.busy = true
	defer func() { c.busy = false }()

	filename := fmt.Sprintf("IMG_%s.jpg", time.Now().Format("20060102_150405"))
	fullPath := filepath.Join(c.dataDir, "original", filename)

	// Ensure dir exists
	os.MkdirAll(filepath.Dir(fullPath), 0755)

	if c.config.Mock {
		return c.mockCapture(fullPath, filename)
	}

	// Real capture
	// gphoto2 --capture-image-and-download --force-overwrite --filename ...
	cmd := exec.Command("gphoto2", "--capture-image-and-download", "--force-overwrite", "--filename", fullPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ Capture failed: %v\nOutput: %s", err, output)
		return "", fmt.Errorf("capture failed: %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found after capture")
	}

	log.Printf("📸 Captured: %s", filename)
	return filename, nil
}

func (c *Controller) mockCapture(fullPath string, filename string) (string, error) {
	log.Printf("📸 [MOCK] Capturing to %s", fullPath)
	time.Sleep(1 * time.Second) // Simulate delay

	// Copy a placeholder image or create a dummy file
	// For now, just create a dummy text file renamed to jpg for testing flow logic
	// In real mock, we might want to generate a real JPEG or fail if we want to test image processing failures.
	// Let's create a minimal valid JPEG header or copy a source asset if we had one.
	// We'll write "MOCK IMAGE CONTENT" for now, image processing will likely fail or warn.
	// Better: Write a 1x1 black pixel JPEG if possible, or just accept that processing might fail on mock data if we don't use a real lib.
	// Since we are using "disintegration/imaging", it expects real image.
	// Let's create a very simple "valid" file if possible, or just fail image processing gracefully.

	// Minimal JPEG header? No, too complex to inline.
	// We will leave it as a text file for now, and handle error in image processor if it fails to decode.
	// OR: We can use the 'imaging' library to create a solid color image and save it!
	// But we need to import it. Let's do that in a separate step or just assume the 'imaging' lib is available since it is in go.mod.

	// Using dummy file for now.
	err := os.WriteFile(fullPath, []byte("MOCK_IMAGE_DATA"), 0644)
	if err != nil {
		return "", err
	}

	return filename, nil
}

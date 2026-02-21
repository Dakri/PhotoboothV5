package camera

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"photobooth/internal/config"
	"photobooth/internal/logging"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CameraInfo holds information about the connected camera.
type CameraInfo struct {
	Connected      bool   `json:"connected"`
	Model          string `json:"model"`
	Manufacturer   string `json:"manufacturer"`
	SerialNumber   string `json:"serialNumber"`
	LensName       string `json:"lensName"`
	BatteryLevel   string `json:"batteryLevel"`
	BatteryPercent int    `json:"batteryPercent"`
	StorageTotal   string `json:"storageTotal"`
	StorageFree    string `json:"storageFree"`
	StoragePercent int    `json:"storagePercent"`
}

// CameraFile represents a file stored on the camera.
type CameraFile struct {
	Name string `json:"name"`
	Size int64  `json:"size"` // Size in KB
}

type Controller struct {
	mu       sync.Mutex
	busy     bool
	config   config.CameraConfig
	dataDir  string
	strategy string // A, B, C, D
	log      *logging.Logger

	// Cached camera info
	infoMu      sync.Mutex
	cachedInfo  CameraInfo
	lastRefresh time.Time
}

func NewController(cfg config.CameraConfig, dataDir string) *Controller {
	return &Controller{
		config:   cfg,
		dataDir:  dataDir,
		strategy: "A",
		log:      logging.Get(),
	}
}

// SetStrategy configures which gphoto2 capture strategy to use (A/B/C/D).
func (c *Controller) SetStrategy(s string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if s == "" {
		s = "A"
	}
	c.strategy = strings.ToUpper(s)
}

// SetDataDir updates the data directory (used when switching albums).
func (c *Controller) SetDataDir(dir string) {
	c.dataDir = dir
}

// IsBusy returns true if the camera is currently capturing.
func (c *Controller) IsBusy() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.busy
}

// GetCachedInfo returns the last cached CameraInfo without touching USB.
func (c *Controller) GetCachedInfo() CameraInfo {
	c.infoMu.Lock()
	defer c.infoMu.Unlock()
	return c.cachedInfo
}

// IsConnected returns true if the camera is currently connected.
func (c *Controller) IsConnected() bool {
	c.infoMu.Lock()
	defer c.infoMu.Unlock()
	return c.cachedInfo.Connected
}

// VerifyLastCapture checks if the most recent photo on the camera has a RAW file.
// Useful to verify if RAW backup was saved to SD card.
func (c *Controller) VerifyLastCapture() (bool, error) {
	if c.config.Mock {
		return true, nil
	}

	c.log.Info("camera", "Verifying if RAW backup exists for last capture on SD card...")

	// List files in folder
	out, err := exec.Command("gphoto2", "--list-files").CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("list-files failed: %v", err)
	}

	output := string(out)
	var latestBase string
	maxNum := -1

	// Find the highest file number to identify the latest captured photo
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		numStr := strings.TrimPrefix(parts[0], "#")
		n := atoi(numStr)
		if n > maxNum {
			maxNum = n
			latestBase = strings.TrimSuffix(parts[1], filepath.Ext(parts[1]))
		}
	}

	if latestBase == "" {
		return false, fmt.Errorf("no files found on camera to verify")
	}

	// Check if any file with the same base name has a RAW extension
	rawExtensions := []string{".arw", ".cr2", ".cr3", ".nef", ".dng", ".raf", ".orf", ".rw2"}
	hasRaw := false

	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, latestBase) {
			// e.g. "#2 IMG_0001.CR2 12345 KB"
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				camFilename := parts[1]
				ext := strings.ToLower(filepath.Ext(camFilename))
				for _, r := range rawExtensions {
					if ext == r {
						hasRaw = true
						break
					}
				}
			}
		}
	}

	if hasRaw {
		c.log.Info("camera", "Verification successful: Found RAW backup for %s on camera", latestBase)
		return true, nil
	}

	c.log.Warn("camera", "CRITICAL Verification failed: RAW backup for %s not found on camera", latestBase)
	return false, nil
}

// DownloadLatestRaw finds the latest RAW file on the camera and downloads it to the given album directory.
func (c *Controller) DownloadLatestRaw(albumDir string) error {
	if c.config.Mock {
		c.log.Info("camera", "[MOCK] Downloading mock RAW file...")
		return nil
	}

	c.log.Info("camera", "Downloading latest RAW file...")

	// List files in folder
	out, err := exec.Command("gphoto2", "--list-files").CombinedOutput()
	if err != nil {
		return fmt.Errorf("list-files failed: %v", err)
	}

	output := string(out)
	var latestRawNum int = -1
	var latestRawName string

	rawExtensions := []string{".arw", ".cr2", ".cr3", ".nef", ".dng", ".raf", ".orf", ".rw2"}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		numStr := strings.TrimPrefix(parts[0], "#")
		n := atoi(numStr)

		ext := strings.ToLower(filepath.Ext(parts[1]))
		isRaw := false
		for _, r := range rawExtensions {
			if ext == r {
				isRaw = true
				break
			}
		}

		if isRaw && n > latestRawNum {
			latestRawNum = n
			latestRawName = parts[1]
		}
	}

	if latestRawNum == -1 {
		return fmt.Errorf("no RAW file found on camera")
	}

	destPath := filepath.Join(albumDir, "original", latestRawName)

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(destPath), 0755)

	// Download the specific RAW file
	dlOut, err := exec.Command("gphoto2", "--get-file", fmt.Sprintf("%d", latestRawNum), "--force-overwrite", "--filename", destPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("get-file failed: %v – %s", err, strings.TrimSpace(string(dlOut)))
	}

	c.log.Info("camera", "Successfully downloaded RAW: %s to %s", latestRawName, destPath)
	return nil
}

// DownloadAllRawToPath downloads all RAW files from the camera to the specified directory.
func (c *Controller) DownloadAllRawToPath(destPath string, onProgress func(copied, total int)) error {
	if c.config.Mock {
		c.log.Info("camera", "[MOCK] Downloading mock RAW files...")
		if onProgress != nil {
			onProgress(1, 1)
		}
		return nil
	}

	c.log.Info("camera", "Listing files for RAW download to %s", destPath)

	out, err := exec.Command("gphoto2", "--list-files").CombinedOutput()
	if err != nil {
		return fmt.Errorf("list-files failed: %v", err)
	}

	output := string(out)
	var rawFileNums []int
	var rawFileNames []string
	rawExtensions := []string{".arw", ".cr2", ".cr3", ".nef", ".dng", ".raf", ".orf", ".rw2"}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		numStr := strings.TrimPrefix(parts[0], "#")
		n := atoi(numStr)
		name := parts[1]

		ext := strings.ToLower(filepath.Ext(name))
		for _, r := range rawExtensions {
			if ext == r {
				rawFileNums = append(rawFileNums, n)
				rawFileNames = append(rawFileNames, name)
				break
			}
		}
	}

	total := len(rawFileNums)
	if total == 0 {
		c.log.Info("camera", "No RAW files found on camera to download.")
		if onProgress != nil {
			onProgress(0, 0) // signal completion
		}
		return nil
	}

	os.MkdirAll(destPath, 0755)

	for i, num := range rawFileNums {
		name := rawFileNames[i]
		targetFile := filepath.Join(destPath, name)

		c.log.Debug("camera", "Downloading RAW %d/%d: %s", i+1, total, name)
		dlOut, err := exec.Command("gphoto2", "--get-file", fmt.Sprintf("%d", num), "--force-overwrite", "--filename", targetFile).CombinedOutput()
		if err != nil {
			c.log.Warn("camera", "Failed to download %s: %v - %s", name, err, string(dlOut))
			// Continue with others even if one fails
		}

		if onProgress != nil {
			onProgress(i+1, total)
		}
	}

	c.log.Info("camera", "Finished downloading %d RAW files.", total)
	return nil
}

// RefreshInfo queries the camera and updates the cache. Only call when idle!
func (c *Controller) RefreshInfo() CameraInfo {
	info := c.queryInfo()
	c.infoMu.Lock()
	c.cachedInfo = info
	c.lastRefresh = time.Now()
	c.infoMu.Unlock()
	return info
}

// ListCameraFiles returns a list of files currently on the camera's storage.
func (c *Controller) ListCameraFiles() ([]CameraFile, error) {
	if c.config.Mock {
		return []CameraFile{
			{Name: "IMG_0001.JPG", Size: 4500},
			{Name: "IMG_0001.CR2", Size: 24500},
			{Name: "IMG_0002.JPG", Size: 4200},
			{Name: "IMG_0002.CR2", Size: 25100},
		}, nil
	}

	c.log.Info("camera", "Listing files on camera...")
	out, err := exec.Command("gphoto2", "--list-files").CombinedOutput()
	if err != nil {
		c.log.Warn("camera", "Failed to list files: %v", err)
		return nil, fmt.Errorf("list-files failed: %v", err)
	}

	var files []CameraFile
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") {
			continue
		}

		// Expected format: "#1     IMG_0001.CR2               12345 KB  image/x-canon-cr2"
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		name := parts[1]
		sizeKB := int64(0)

		// Find "KB" and parse the number before it
		for i, p := range parts {
			if strings.ToUpper(p) == "KB" && i > 0 {
				sizeKB = int64(atoi(parts[i-1]))
				break
			}
		}

		files = append(files, CameraFile{
			Name: name,
			Size: sizeKB,
		})
	}

	return files, nil
}

// queryInfo does the actual gphoto2 queries.
func (c *Controller) queryInfo() CameraInfo {
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
	c.killGphotoBlockers()

	summaryOut, err := exec.Command("gphoto2", "--summary").CombinedOutput()
	if err != nil {
		c.log.Warn("camera", "No camera detected: %v", err)
		return info
	}

	rawSummary := strings.TrimSpace(string(summaryOut))
	info.Connected = true
	parseSummary(rawSummary, &info)
	c.log.Info("camera", "Camera: %s | Lens: %s | Battery: %s", info.Model, info.LensName, info.BatteryLevel)

	storageOut, err := exec.Command("gphoto2", "--storage-info").CombinedOutput()
	if err == nil {
		parseStorage(strings.TrimSpace(string(storageOut)), &info)
		if info.StorageFree != "" {
			c.log.Info("camera", "Storage: %s free / %s total", info.StorageFree, info.StorageTotal)
		}
	}

	return info
}

// PrepareCapture is called when countdown starts to pre-configure the camera.
func (c *Controller) PrepareCapture() {
	if c.config.Mock {
		return
	}
	// Run in background to avoid blocking the countdown
	go func() {
		target := "1" // Default: Memory Card (Strategies A, B, D)
		c.mu.Lock()
		strategy := strings.ToUpper(c.strategy)
		c.mu.Unlock()

		if strategy == "C" {
			target = "0" // Internal RAM for Strategy C (No SD backup)
		}

		// Set capturetarget
		// This can take ~200-500ms, so doing it during countdown saves time at capture.
		if out, err := exec.Command("gphoto2", "--set-config", fmt.Sprintf("capturetarget=%s", target)).CombinedOutput(); err != nil {
			c.log.Warn("camera", "PrepareCapture: failed to set capturetarget=%s: %v – %s", target, err, strings.TrimSpace(string(out)))
		} else {
			c.log.Debug("camera", "PrepareCapture: capturetarget=%s set", target)
		}
	}()
}

// ─────────────────────────────────────────────────────────────────────────────
// Capture
// ─────────────────────────────────────────────────────────────────────────────

// Capture runs all 4 strategies timed, picks the first one that succeeds,
// logs detailed benchmark results, and returns the filename.
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
	os.MkdirAll(filepath.Dir(fullPath), 0755)

	if c.config.Mock {
		return c.mockCapture(fullPath, filename)
	}

	// Dispatch to selected strategy
	strategy := strings.ToUpper(c.strategy)
	if strategy == "" {
		strategy = "A"
	}
	c.log.Info("camera", "Using capture strategy %s", strategy)

	var dur time.Duration
	var err error
	switch strategy {
	case "A":
		dur, err = c.strategySDTargetGetFile(fullPath)
	case "B":
		dur, err = c.strategyDownloadAllSaveLocally(fullPath)
	case "C":
		dur, err = c.strategyDownloadAllRemoveFromSD(fullPath)
	case "D":
		dur, err = c.strategyTethered(fullPath)
	default:
		c.log.Warn("camera", "Unknown strategy %q – falling back to A", strategy)
		dur, err = c.strategySDTargetGetFile(fullPath)
	}

	if err != nil {
		c.log.Error("camera", "Strategy %s failed after %.3fs: %v", strategy, dur.Seconds(), err)
		return "", err
	}

	stat, _ := os.Stat(fullPath)
	sizeKB := int64(0)
	if stat != nil {
		sizeKB = stat.Size() / 1024
	}
	c.log.Info("camera", "Capture done [%s] %.3fs – %s (%d KB)", strategy, dur.Seconds(), filename, sizeKB)
	return filename, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Strategy A: SD-Target + list-files + get-file
// RAW stays on SD card, we download only the JPEG via --get-file
// ─────────────────────────────────────────────────────────────────────────────
func (c *Controller) strategySDTargetGetFile(destPath string) (time.Duration, error) {
	t0 := time.Now()

	// target=1 already set by PrepareCapture during countdown

	// Fire camera and let it save to SD card
	c.log.Info("camera", "  A: Capturing to SD card...")

	// Trigger shutter (no download)
	tShutter := time.Now()
	if out, err := exec.Command("gphoto2", "--capture-image").CombinedOutput(); err != nil {
		return time.Since(t0), fmt.Errorf("capture-image failed: %v – %s", err, strings.TrimSpace(string(out)))
	}
	c.log.Info("benchmark", "  A: Shutter %.3fs", time.Since(tShutter).Seconds())

	// List files and find newest JPEG
	tList := time.Now()
	listOut, err := exec.Command("gphoto2", "--list-files").CombinedOutput()
	if err != nil {
		return time.Since(t0), fmt.Errorf("list-files failed: %v", err)
	}
	c.log.Info("benchmark", "  A: ListFiles %.3fs", time.Since(tList).Seconds())

	jpegNum := findLatestJPEGNum(string(listOut))
	if jpegNum < 0 {
		return time.Since(t0), fmt.Errorf("no JPEG found in file list")
	}

	// Download only the JPEG
	tDl := time.Now()
	dlOut, err := exec.Command("gphoto2", "--get-file", fmt.Sprintf("%d", jpegNum),
		"--force-overwrite", "--filename", destPath).CombinedOutput()
	if err != nil {
		return time.Since(t0), fmt.Errorf("get-file failed: %v – %s", err, strings.TrimSpace(string(dlOut)))
	}
	c.log.Info("benchmark", "  A: Download %.3fs | Total %.3fs", time.Since(tDl).Seconds(), time.Since(t0).Seconds())

	return time.Since(t0), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Strategy B: Download-All → Save everything locally (RAW + JPEG)
// captures to SD (if supported) AND downloads everything to the Pi.
// RAWs are saved in the same folder as the JPEG.
// ─────────────────────────────────────────────────────────────────────────────
func (c *Controller) strategyDownloadAllSaveLocally(destPath string) (time.Duration, error) {
	t0 := time.Now()

	// capturetarget=1 is now set during countdown via PrepareCapture()

	tmpDir, err := os.MkdirTemp("", "pb-b-")
	if err != nil {
		return 0, fmt.Errorf("tmp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c.log.Info("camera", "  B: Capturing & downloading all files...")
	tCapture := time.Now()
	// Download all files to temp dir
	// Added --keep-raw as requested to try keeping RAW on camera
	out, err := exec.Command("gphoto2", "--capture-image-and-download", "--keep-raw", "--force-overwrite",
		"--filename", filepath.Join(tmpDir, "%f.%C")).CombinedOutput()
	if err != nil {
		return time.Since(t0), fmt.Errorf("capture-and-download failed: %v – %s", err, strings.TrimSpace(string(out)))
	}
	c.log.Info("camera", "  B: Capture+Download %.3fs", time.Since(tCapture).Seconds())

	entries, _ := os.ReadDir(tmpDir)
	var jpegSrc string
	savedFiles := []string{}

	baseName := strings.TrimSuffix(filepath.Base(destPath), filepath.Ext(destPath))
	destDir := filepath.Dir(destPath)

	for _, e := range entries {
		src := filepath.Join(tmpDir, e.Name())
		ext := filepath.Ext(e.Name())
		lowerExt := strings.ToLower(ext)

		// Determine new filename: IMG_YYYYMMDD_HHMMSS.<ext>
		destFile := filepath.Join(destDir, baseName+ext)

		if lowerExt == ".jpg" || lowerExt == ".jpeg" {
			jpegSrc = src
			// Move JPEG to exact destPath requested by controller (to ensure name match)
			err = copyFile(src, destPath)
		} else {
			// Move RAW/other files to same dir with same basename
			err = copyFile(src, destFile)
		}

		if err != nil {
			c.log.Warn("camera", "  B: Failed to move %s: %v", e.Name(), err)
		} else {
			savedFiles = append(savedFiles, filepath.Base(destFile))
		}
	}

	if jpegSrc == "" {
		return time.Since(t0), fmt.Errorf("no JPEG in download")
	}

	c.log.Info("camera", "  B: Saved locally: %v", savedFiles)
	return time.Since(t0), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Strategy C: Download-All (No SD Backup)
// Sets capturetarget=0 (RAM) or captures and omits --keep-raw so nothing stays on SD.
// Both JPEG and RAW are downloaded to the Raspberry Pi.
// ─────────────────────────────────────────────────────────────────────────────
func (c *Controller) strategyDownloadAllRemoveFromSD(destPath string) (time.Duration, error) {
	t0 := time.Now()

	tmpDir, err := os.MkdirTemp("", "pb-c-")
	if err != nil {
		return 0, fmt.Errorf("tmp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	c.log.Info("camera", "  C: Capturing & downloading all files (No SD backup)...")
	tCapture := time.Now()
	// Download all files to temp dir.
	// OMIT --keep-raw so that files are deleted from the camera after download.
	out, err := exec.Command("gphoto2", "--capture-image-and-download", "--force-overwrite",
		"--filename", filepath.Join(tmpDir, "%f.%C")).CombinedOutput()
	if err != nil {
		return time.Since(t0), fmt.Errorf("capture-and-download failed: %v – %s", err, strings.TrimSpace(string(out)))
	}
	c.log.Info("camera", "  C: Capture+Download %.3fs", time.Since(tCapture).Seconds())

	entries, _ := os.ReadDir(tmpDir)
	baseName := strings.TrimSuffix(filepath.Base(destPath), filepath.Ext(destPath))
	destDir := filepath.Dir(destPath)

	for _, e := range entries {
		src := filepath.Join(tmpDir, e.Name())
		ext := filepath.Ext(e.Name())
		lowerExt := strings.ToLower(ext)

		destFile := filepath.Join(destDir, baseName+ext)

		if lowerExt == ".jpg" || lowerExt == ".jpeg" {
			err = copyFile(src, destPath)
		} else {
			err = copyFile(src, destFile)
		}

		if err != nil {
			c.log.Warn("camera", "  C: Failed to copy %s: %v", e.Name(), err)
		}
	}

	return time.Since(t0), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Strategy D: Tethered capture (experimental)
// gphoto2 --capture-tethered waits for a shutter event from the camera itself
// or a trigger. We run it with a timeout and grab whatever it downloads.
// This is useful for remote-trigger workflows and sometimes faster because
// gphoto2 starts the USB transfer immediately when the camera signals "done".
// ─────────────────────────────────────────────────────────────────────────────
func (c *Controller) strategyTethered(destPath string) (time.Duration, error) {
	t0 := time.Now()

	tmpDir, err := os.MkdirTemp("", "pb-d-")
	if err != nil {
		return 0, fmt.Errorf("tmp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Start tethered session – gphoto2 will trigger the shutter and receive files
	cmd := exec.Command("gphoto2", "--capture-tethered",
		"--hook-script=/dev/null", // avoid hook errors
		"--frames=1",              // only capture one frame
		"--interval=0",            // immediately
		"--force-overwrite",
		"--filename", filepath.Join(tmpDir, "%f.%C"))

	if err := cmd.Start(); err != nil {
		return time.Since(t0), fmt.Errorf("tethered start failed: %v", err)
	}

	// Wait max 30 seconds for a frame
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		if err != nil {
			return time.Since(t0), fmt.Errorf("tethered failed: %v", err)
		}
	case <-time.After(30 * time.Second):
		cmd.Process.Kill()
		return time.Since(t0), fmt.Errorf("tethered timed out after 30s")
	}

	c.log.Info("benchmark", "  D: Tethered trigger+download %.3fs", time.Since(t0).Seconds())

	// Find JPEG in tmp dir
	entries, _ := os.ReadDir(tmpDir)
	var jpegSrc string
	for _, e := range entries {
		if ext := strings.ToLower(filepath.Ext(e.Name())); ext == ".jpg" || ext == ".jpeg" {
			jpegSrc = filepath.Join(tmpDir, e.Name())
		}
	}
	if jpegSrc == "" {
		return time.Since(t0), fmt.Errorf("tethered: no JPEG received")
	}

	if err := copyFile(jpegSrc, destPath); err != nil {
		return time.Since(t0), fmt.Errorf("tethered: copy failed: %v", err)
	}

	c.log.Info("benchmark", "  D: Total %.3fs", time.Since(t0).Seconds())
	return time.Since(t0), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func (c *Controller) mockCapture(fullPath string, filename string) (string, error) {
	c.log.Info("camera", "[MOCK] Capturing to %s", fullPath)
	time.Sleep(1 * time.Second)

	jpegBytes := []byte{
		0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01,
		0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xFF, 0xDB, 0x00, 0x43,
		0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08, 0x07, 0x07, 0x07, 0x09,
		0x09, 0x08, 0x0A, 0x0C, 0x14, 0x0D, 0x0C, 0x0B, 0x0B, 0x0C, 0x19, 0x12,
		0x13, 0x0F, 0x14, 0x1D, 0x1A, 0x1F, 0x1E, 0x1D, 0x1A, 0x1C, 0x1C, 0x20,
		0x24, 0x2E, 0x27, 0x20, 0x22, 0x2C, 0x23, 0x1C, 0x1C, 0x28, 0x37, 0x29,
		0x2C, 0x30, 0x31, 0x34, 0x34, 0x34, 0x1F, 0x27, 0x39, 0x3D, 0x38, 0x32,
		0x3C, 0x2E, 0x33, 0x34, 0x32, 0xFF, 0xC0, 0x00, 0x0B, 0x08, 0x00, 0x01,
		0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xFF, 0xC4, 0x00, 0x1F, 0x00, 0x00,
		0x01, 0x05, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0xFF, 0xDA, 0x00, 0x08, 0x01, 0x01, 0x00, 0x00, 0x3F,
		0x00, 0x7B, 0x94, 0x11, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xD9,
	}

	return filename, os.WriteFile(fullPath, jpegBytes, 0644)
}

func (c *Controller) killGphotoBlockers() {
	killed := false
	if _, err := exec.Command("pkill", "-f", "gvfsd-gphoto2").CombinedOutput(); err == nil {
		c.log.Info("camera", "Killed gvfsd-gphoto2")
		killed = true
	}
	if _, err := exec.Command("pkill", "-f", "gvfs-gphoto2-volume-monitor").CombinedOutput(); err == nil {
		c.log.Info("camera", "Killed gvfs-gphoto2-volume-monitor")
		killed = true
	}
	if killed {
		time.Sleep(500 * time.Millisecond)
	}
}

// findLatestJPEGNum parses gphoto2 --list-files output and returns the highest file number for a JPEG.
func findLatestJPEGNum(output string) int {
	best := -1
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		numStr := strings.TrimPrefix(parts[0], "#")
		ext := strings.ToLower(filepath.Ext(parts[1]))
		if ext == ".jpg" || ext == ".jpeg" {
			if n := atoi(numStr); n > best {
				best = n
			}
		}
	}
	return best
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return -1
	}
	return n
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
			// Try to parse percent
			if strings.Contains(val, "%") {
				parts := strings.Split(val, "%")
				if n, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
					info.BatteryPercent = n
				}
			} else {
				// Some cameras might return "High" or similar
				if val == "High" {
					info.BatteryPercent = 100
				}
				if val == "Medium" {
					info.BatteryPercent = 50
				}
				if val == "Low" {
					info.BatteryPercent = 10
				}
			}
		}
	}
}

// parseStorage extracts storage capacity from gphoto2 --storage-info output.
func parseStorage(output string, info *CameraInfo) {
	var totalBytes, freeBytes int64

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		var key, val string
		if strings.Contains(line, "=") {
			p := strings.SplitN(line, "=", 2)
			key, val = p[0], p[1]
		} else if strings.Contains(line, ":") {
			p := strings.SplitN(line, ":", 2)
			key, val = p[0], p[1]
		} else {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		val = strings.TrimSpace(val)

		switch key {
		case "totalcapacity", "capacity", "total capacity":
			totalBytes = parseBytes(val)
		case "free", "free space", "freespace":
			freeBytes = parseBytes(val)
		}
	}

	if totalBytes > 0 {
		info.StorageTotal = fmt.Sprintf("%.1f GB", float64(totalBytes)/1024/1024/1024)
		info.StorageFree = fmt.Sprintf("%.1f GB", float64(freeBytes)/1024/1024/1024)
		if freeBytes > 0 {
			info.StoragePercent = int((float64(freeBytes) / float64(totalBytes)) * 100)
		}
	} else {
		// Fallback for mock or weird output
		info.StorageTotal = "?"
		info.StorageFree = "?"
	}
}

func parseBytes(s string) int64 {
	s = strings.TrimSpace(strings.ToUpper(s))
	// Remove "Available: " etc if present
	if idx := strings.LastIndex(s, " "); idx != -1 {
		// Check if last part is unit
		// s might be "32768 KB"
		// or "capacity=32768" (no unit, assume KB from gphoto default?)
		// actually gphoto usually gives "Total Capacity=31899648" (KB usually)
	}

	// Regex would be safer but let's try manual parsing standard formats
	// "1234 KB", "1234 MB", "1234" (KB default)

	unit := "KB" // Default assumption for gphoto
	valStr := s

	if strings.HasSuffix(s, "GB") {
		unit = "GB"
		valStr = strings.TrimSuffix(s, "GB")
	} else if strings.HasSuffix(s, "MB") {
		unit = "MB"
		valStr = strings.TrimSuffix(s, "MB")
	} else if strings.HasSuffix(s, "KB") {
		unit = "KB"
		valStr = strings.TrimSuffix(s, "KB")
	} else if strings.HasSuffix(s, "B") {
		unit = "B"
		valStr = strings.TrimSuffix(s, "B")
	}

	val, err := strconv.ParseInt(strings.TrimSpace(valStr), 10, 64)
	if err != nil {
		return 0
	}

	switch unit {
	case "GB":
		return val * 1024 * 1024 * 1024
	case "MB":
		return val * 1024 * 1024
	case "KB":
		return val * 1024
	default:
		return val
	}
}

package imaging

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"photobooth/internal/config"
	"photobooth/internal/logging"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
)

type Processor struct {
	config  config.ImageConfig
	log     *logging.Logger
	useEpeg bool
}

func NewProcessor(cfg config.ImageConfig) *Processor {
	p := &Processor{
		config: cfg,
		log:    logging.Get(),
	}

	// Check if epeg is available
	if path, err := exec.LookPath("epeg"); err == nil && path != "" {
		p.useEpeg = true
		p.log.Info("imaging", "Using 'epeg' for fast JPEG processing")
	} else {
		p.log.Info("imaging", "'epeg' not found – using native Go imaging (slower)")
	}

	return p
}

func (p *Processor) Process(originalPath string, onPreviewReady func()) error {
	start := time.Now()
	filename := filepath.Base(originalPath)
	baseDir := filepath.Dir(filepath.Dir(originalPath)) // data/photos
	previewPath := filepath.Join(baseDir, "preview", filename)
	thumbPath := filepath.Join(baseDir, "thumb", filename)

	var wg sync.WaitGroup
	wg.Add(2)

	// Function to generate image (either via epeg or go-native)
	generate := func(destPath string, width int, quality int) error {
		defer wg.Done()

		if p.useEpeg {
			// Try epeg first
			// Use -m (max dimension) instead of -w to handle portrait/landscape better
			cmd := exec.Command("epeg",
				"-m", fmt.Sprintf("%d", width),
				"-q", fmt.Sprintf("%d", quality),
				originalPath, destPath)

			if out, err := cmd.CombinedOutput(); err == nil {
				// Success? Check file size to catch "solid color" bug
				if info, err := os.Stat(destPath); err == nil && info.Size() > 3000 {
					return nil // EPEG worked and produced a reasonable file
				} else {
					p.log.Warn("imaging", "EPEG produced suspicious file (size=%d), falling back to Go", info.Size())
					// Proceed to fallback...
				}
			} else {
				p.log.Warn("imaging", "EPEG failed: %v – %s", err, strings.TrimSpace(string(out)))
				// Proceed to fallback...
			}
		}

		// Fallback: Go native
		// Re-open source for each thread to be safe/simple
		src, err := imaging.Open(originalPath, imaging.AutoOrientation(true))
		if err != nil {
			return err
		}

		// Resize (using Fit to match -m max dimension behavior)
		dst := imaging.Fit(src, width, width, imaging.Lanczos)
		return imaging.Save(dst, destPath, imaging.JPEGQuality(quality))
	}

	// Generate Preview
	go func() {
		err := generate(previewPath, p.config.PreviewWidth, p.config.PreviewQuality)
		if err != nil {
			p.log.Error("imaging", "Failed to generate preview: %v", err)
			// Try fallback if epeg failed?
		} else if onPreviewReady != nil {
			onPreviewReady()
		}
	}()

	// Generate Thumbnail
	go func() {
		err := generate(thumbPath, p.config.ThumbWidth, p.config.ThumbQuality)
		if err != nil {
			p.log.Error("imaging", "Failed to generate thumbnail: %v", err)
		}
	}()

	wg.Wait()
	p.log.Info("imaging", "Processed %s in %v (epeg=%v)", filename, time.Since(start).Round(time.Millisecond), p.useEpeg)
	return nil
}

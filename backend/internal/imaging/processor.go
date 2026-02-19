package imaging

import (
	"log"
	"path/filepath"
	"photobooth/internal/config"
	"sync"
	"time"

	"github.com/disintegration/imaging"
)

type Processor struct {
	config config.ImageConfig
}

func NewProcessor(cfg config.ImageConfig) *Processor {
	return &Processor{
		config: cfg,
	}
}

func (p *Processor) Process(originalPath string) error {
	start := time.Now()

	// Open src image
	src, err := imaging.Open(originalPath, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	dir := filepath.Dir(originalPath)
	filename := filepath.Base(originalPath)

	// Paths
	// Assuming originalPath is .../data/photos/original/IMG.jpg
	// We want .../data/photos/preview/IMG.jpg
	// and .../data/photos/thumb/IMG.jpg

	// Hacky path manipulation based on known structure, better to pass base dir
	// But let's assume standard structure: data/photos/{original,preview,thumb}
	baseDir := filepath.Dir(dir) // data/photos
	previewPath := filepath.Join(baseDir, "preview", filename)
	thumbPath := filepath.Join(baseDir, "thumb", filename)

	var wg sync.WaitGroup
	wg.Add(2)

	// Generate Preview
	go func() {
		defer wg.Done()
		dst := imaging.Fit(src, p.config.PreviewWidth, p.config.PreviewWidth, imaging.Lanczos)
		if err := imaging.Save(dst, previewPath, imaging.JPEGQuality(80)); err != nil {
			log.Printf("❌ Failed to save preview: %v", err)
		}
	}()

	// Generate Thumbnail
	go func() {
		defer wg.Done()
		dst := imaging.Thumbnail(src, p.config.ThumbWidth, p.config.ThumbWidth, imaging.Lanczos)
		if err := imaging.Save(dst, thumbPath, imaging.JPEGQuality(70)); err != nil {
			log.Printf("❌ Failed to save thumbnail: %v", err)
		}
	}()

	wg.Wait()
	log.Printf("✅ Processed image in %v", time.Since(start))
	return nil
}

package storage

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Photo struct {
	Filename  string    `json:"filename"`
	Timestamp time.Time `json:"timestamp"`
	Url       string    `json:"url"` // Preview URL
	ThumbUrl  string    `json:"thumbUrl"`
}

type Manager struct {
	rootDir string // data/photos
}

func NewManager(rootDir string) *Manager {
	return &Manager{rootDir: rootDir}
}

// SetRootDir updates the root directory (used when switching albums).
func (m *Manager) SetRootDir(dir string) {
	m.rootDir = dir
}

func (m *Manager) EnsureDirs() {
	dirs := []string{"original", "preview", "thumb"}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(m.rootDir, d), 0755)
	}
}

func (m *Manager) List() ([]Photo, error) {
	dir := filepath.Join(m.rootDir, "original")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var photos []Photo
	for _, e := range entries {
		if !e.IsDir() && isImage(e.Name()) {
			info, err := e.Info()
			if err != nil {
				continue
			}
			photos = append(photos, Photo{
				Filename:  e.Name(),
				Timestamp: info.ModTime(),
				Url:       "/photos/preview/" + e.Name(),
				ThumbUrl:  "/photos/thumb/" + e.Name(),
			})
		}
	}

	// Sort newest first
	sort.Slice(photos, func(i, j int) bool {
		return photos[i].Timestamp.After(photos[j].Timestamp)
	})

	return photos, nil
}

func isImage(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}

func (m *Manager) GetLatest() *Photo {
	photos, _ := m.List()
	if len(photos) > 0 {
		return &photos[0]
	}
	return nil
}

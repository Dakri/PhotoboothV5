package disk

import "fmt"

// Usage holds disk usage information.
type Usage struct {
	Total       uint64 `json:"total"`
	Free        uint64 `json:"free"`
	Used        uint64 `json:"used"`
	UsedPercent int    `json:"usedPercent"`
}

// GetUsage returns disk usage for the file system containing path.
func GetUsage(path string) (Usage, error) {
	return getUsage(path)
}

func (u Usage) String() string {
	return fmt.Sprintf("Total: %d, Free: %d, Used: %d (%d%%)", u.Total, u.Free, u.Used, u.UsedPercent)
}

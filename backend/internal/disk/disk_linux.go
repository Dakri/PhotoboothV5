package disk

import (
	"syscall"
)

func getUsage(path string) (Usage, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return Usage{}, err
	}

	// Blocks * BlockSize = Bytes
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free

	var percent int
	if total > 0 {
		percent = int((float64(used) / float64(total)) * 100)
	}

	return Usage{
		Total:       total,
		Free:        free,
		Used:        used,
		UsedPercent: percent,
	}, nil
}

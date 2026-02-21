package disk

func getUsage(path string) (Usage, error) {
	// Stub for Windows development
	// Return some mock data or 0
	return Usage{
		Total:       1024 * 1024 * 1024 * 64, // 64 GB
		Free:        1024 * 1024 * 1024 * 40, // 40 GB
		Used:        1024 * 1024 * 1024 * 24, // 24 GB
		UsedPercent: 37,
	}, nil
}

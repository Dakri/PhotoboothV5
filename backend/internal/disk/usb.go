package disk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type UsbDevice struct {
	Name       string `json:"name"`       // sda1
	Label      string `json:"label"`      // KINGSTON
	MountPoint string `json:"mountpoint"` // /media/usb
	Size       string `json:"size"`       // 32G
	Free       string `json:"free"`       // 15G
	Subsystems string `json:"subsystems"` // block:scsi:usb:pci
}

// LsblkOutput matches the JSON output of `lsblk -J -O` (or similar).
type LsblkOutput struct {
	Blockdevices []struct {
		Name       string `json:"name"`
		Label      string `json:"label"`
		Mountpoint string `json:"mountpoint"`
		Size       string `json:"size"`
		Fsavail    string `json:"fsavail"`
		Subsystems string `json:"subsystems"`
		Children   []struct {
			Name       string `json:"name"`
			Label      string `json:"label"`
			Mountpoint string `json:"mountpoint"`
			Size       string `json:"size"`
			Fsavail    string `json:"fsavail"`
			Subsystems string `json:"subsystems"`
		} `json:"children"`
	} `json:"blockdevices"`
}

// GetUsbDevices uses lsblk to list USB block devices (partitions).
func GetUsbDevices() ([]UsbDevice, error) {
	out, err := exec.Command("lsblk", "-J", "-o", "NAME,LABEL,MOUNTPOINT,SIZE,FSAVAIL,SUBSYSTEMS").Output()
	if err != nil {
		return nil, fmt.Errorf("lsblk failed: %v", err)
	}

	var parsed LsblkOutput
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, fmt.Errorf("unmarshal lsblk failed: %v", err)
	}

	var devices []UsbDevice

	for _, dev := range parsed.Blockdevices {
		// Only look at USB subsystems (or their children)
		isUsb := strings.Contains(dev.Subsystems, "usb")

		for _, child := range dev.Children {
			// Often the parent is the USB device, and the child is the partition we want to mount
			if isUsb || strings.Contains(child.Subsystems, "usb") {
				label := child.Label
				if label == "" {
					label = "USB Laufwerk"
				}
				devices = append(devices, UsbDevice{
					Name:       child.Name,
					Label:      label,
					MountPoint: child.Mountpoint,
					Size:       child.Size,
					Free:       child.Fsavail,
					Subsystems: child.Subsystems,
				})
			}
		}
	}

	return devices, nil
}

// MountUsb mounts a device (e.g., "sda1") to a given directory.
func MountUsb(deviceName string) (string, error) {
	devices, err := GetUsbDevices()
	if err != nil {
		return "", err
	}

	var dev *UsbDevice
	for _, d := range devices {
		if d.Name == deviceName {
			dev = &d
			break
		}
	}

	if dev == nil {
		return "", fmt.Errorf("device %s not found", deviceName)
	}

	if dev.MountPoint != "" {
		return dev.MountPoint, nil // Already mounted
	}

	mountPath := filepath.Join("/media", dev.Name)
	os.MkdirAll(mountPath, 0777)

	cmd := exec.Command("sudo", "mount", "/dev/"+dev.Name, mountPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("mount failed: %v\n%s", err, string(out))
	}

	return mountPath, nil
}

// UnmountUsb flushes the file system buffers and unmounts the USB device.
func UnmountUsb(mountPoint string) error {
	// First flush all file system buffers to ensure data is written to the USB drive
	exec.Command("sync").Run()

	cmd := exec.Command("sudo", "umount", mountPoint)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("umount failed: %v\n%s", err, string(out))
	}
	return nil
}

// CopyDir recursively copies a directory to a destination.
func CopyDir(ctx context.Context, src string, dst string) error {
	return CopyDirWithProgress(ctx, src, dst, nil)
}

// CopyDirWithProgress recursively copies a directory to a destination and reports progress.
func CopyDirWithProgress(ctx context.Context, src string, dst string, onProgress func(copiedBytes, totalBytes, copiedFiles, totalFiles int64)) error {
	var totalBytes, totalFiles, copiedBytes, copiedFiles int64

	filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			totalFiles++
			totalBytes += info.Size()
		}
		return nil
	})

	var copyRecursive func(s, d string) error
	copyRecursive = func(s, d string) error {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		entries, err := os.ReadDir(s)
		if err != nil {
			return err
		}

		os.MkdirAll(d, 0777)

		for _, entry := range entries {
			srcPath := filepath.Join(s, entry.Name())
			dstPath := filepath.Join(d, entry.Name())

			if entry.IsDir() {
				err = copyRecursive(srcPath, dstPath)
				if err != nil {
					return err
				}
			} else {
				// Check for cancellation before each file
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				info, err := entry.Info()
				var size int64
				if err == nil {
					size = info.Size()
				}

				err = copyFile(srcPath, dstPath)
				if err != nil {
					return err
				}

				copiedFiles++
				copiedBytes += size
				if onProgress != nil {
					onProgress(copiedBytes, totalBytes, copiedFiles, totalFiles)
				}
			}
		}
		return nil
	}

	return copyRecursive(src, dst)
}

func copyFile(srcFile, dstFile string) error {
	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	return err
}

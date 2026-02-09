//go:build !linux

package monitor

import "errors"

func CPUPercent() (string, error) {
	return "", errors.New("monitoring is only supported on linux")
}

func MemoryUsage() (string, error) {
	return "", errors.New("monitoring is only supported on linux")
}

func DiskUsage(path string) (string, error) {
	return "", errors.New("monitoring is only supported on linux")
}

func Uptime() (string, error) {
	return "", errors.New("monitoring is only supported on linux")
}

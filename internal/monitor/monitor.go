//go:build linux

package monitor

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func CPUPercent() (string, error) {
	idle1, total1, err := readCPUStat()
	if err != nil {
		return "", err
	}
	time.Sleep(500 * time.Millisecond)
	idle2, total2, err := readCPUStat()
	if err != nil {
		return "", err
	}
	idleTicks := float64(idle2 - idle1)
	totalTicks := float64(total2 - total1)
	if totalTicks == 0 {
		return "0%", nil
	}
	usage := (1.0 - idleTicks/totalTicks) * 100
	return fmt.Sprintf("%.2f%%", usage), nil
}

func MemoryUsage() (string, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return "", err
	}
	defer file.Close()

	var total, available float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			total = parseMemValue(line)
		}
		if strings.HasPrefix(line, "MemAvailable:") {
			available = parseMemValue(line)
		}
	}
	if total == 0 {
		return "", fmt.Errorf("failed to read memory info")
	}
	used := total - available
	return fmt.Sprintf("%.1f%% (%.0fMB/%.0fMB)", used/total*100, used/1024, total/1024), nil
}

func DiskUsage(path string) (string, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return "", err
	}

	total := float64(stat.Blocks) * float64(stat.Bsize)
	free := float64(stat.Bavail) * float64(stat.Bsize)
	used := total - free
	if total == 0 {
		return "0%", nil
	}
	return fmt.Sprintf("%.1f%% (%.1fGB/%.1fGB)", used/total*100, used/1024/1024/1024, total/1024/1024/1024), nil
}

func Uptime() (string, error) {
	content, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "", err
	}
	parts := strings.Fields(string(content))
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid uptime format")
	}
	seconds, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return "", err
	}
	return formatDuration(time.Duration(seconds) * time.Second), nil
}

func readCPUStat() (idle, total uint64, err error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return 0, 0, fmt.Errorf("empty stat")
	}
	fields := strings.Fields(scanner.Text())
	if len(fields) < 5 {
		return 0, 0, fmt.Errorf("invalid stat format")
	}

	var values []uint64
	for _, f := range fields[1:] {
		val, err := strconv.ParseUint(f, 10, 64)
		if err != nil {
			return 0, 0, err
		}
		values = append(values, val)
		total += val
	}
	idle = values[3]
	if len(values) > 4 {
		idle += values[4]
	}
	return idle, total, nil
}

func parseMemValue(line string) float64 {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return 0
	}
	val, _ := strconv.ParseFloat(parts[1], 64)
	return val
}

func formatDuration(d time.Duration) string {
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

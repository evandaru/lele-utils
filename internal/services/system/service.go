package system

import (
	"fmt"
	"runtime"
	"time"

	"dev-utils/internal/models"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// Service — PRD §17.
type Service struct{}

func New() *Service { return &Service{} }

func (s *Service) Info() (*models.SystemInfo, error) {
	info := &models.SystemInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
	if h, err := host.Info(); err == nil {
		info.Kernel = h.KernelVersion
		info.Platform = h.Platform
		info.Uptime = formatUptime(time.Duration(h.Uptime) * time.Second)
		info.UptimeDays = float64(h.Uptime) / 86400
		if h.OS != "" {
			info.OS = prettyOS(h.OS, h.Platform)
		}
	}
	if c, err := cpu.Info(); err == nil && len(c) > 0 {
		info.CPUModel = c[0].ModelName
	}
	if n, err := cpu.Counts(true); err == nil {
		info.CPUCores = n
	}
	if m, err := mem.VirtualMemory(); err == nil {
		info.MemTotalGB = float64(m.Total) / 1e9
		info.MemUsedGB = float64(m.Used) / 1e9
	}
	if d, err := disk.Usage("/"); err == nil {
		info.DiskTotalGB = float64(d.Total) / 1e9
		info.DiskUsedGB = float64(d.Used) / 1e9
	}
	if info.OS == "darwin" {
		info.OS = "macOS"
		if info.Arch == "arm64" {
			// sudah benar
		}
	} else if info.OS == "linux" {
		info.OS = "Linux"
	}
	if info.Kernel == "" {
		info.Kernel = "unknown"
	}
	return info, nil
}

func prettyOS(os, platform string) string {
	if os == "darwin" {
		return "darwin"
	}
	return os
}

func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if days > 0 {
		return fmt.Sprintf("%d days, %d hours", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%d hours", hours)
	}
	mins := int(d.Minutes()) % 60
	return fmt.Sprintf("%d minutes", mins)
}

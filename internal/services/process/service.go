package process

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"syscall"

	"lele-dev/internal/models"
	"lele-dev/internal/platform/darwin"
	"lele-dev/internal/platform/linux"
)

// Service implements platform.ProcessService (PRD §24).
// UI hanya memanggil List/Kill — tidak pernah exec langsung.
type Service struct{}

func New() *Service { return &Service{} }

// List returns all processes (platform-specific listing, satu interface).
func (s *Service) List() ([]models.Process, error) {
	if runtime.GOOS == "darwin" {
		return darwin.ListProcesses()
	}
	return linux.ListProcesses()
}

// Find returns single process detail.
func (s *Service) Find(pid int) (*models.Process, error) {
	all, err := s.List()
	if err != nil {
		return nil, err
	}
	for _, p := range all {
		if p.PID == pid {
			cp := p
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("process dengan PID %d tidak ditemukan", pid)
}

// Search filters by name/command/PID substring (PRD §18: `/`).
func Search(all []models.Process, query string) []models.Process {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return all
	}
	var out []models.Process
	for _, p := range all {
		if strings.Contains(strings.ToLower(p.Name), q) ||
			strings.Contains(strings.ToLower(p.Command), q) ||
			strings.Contains(fmt.Sprint(p.PID), q) {
			out = append(out, p)
		}
	}
	return out
}

// SortByCPU sorts descending CPU (helper TUI).
func SortByCPU(all []models.Process) {
	sort.Slice(all, func(i, j int) bool { return all[i].CPUPercent > all[j].CPUPercent })
}

// IsProtected — PRD §23 rule 4: process system harus dilindungi.
func IsProtected(pid int) bool {
	return pid <= 2 // 0 kernel scheduler, 1 init, 2 kthreadd
}

// CanSignal checks permission tanpa membunuh (signal 0).
func CanSignal(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal 0 = cek permission saja.
	if err := p.Signal(syscall.Signal(0)); err != nil {
		return false
	}
	return true
}

// Kill sends SIGTERM (default, PRD §10). Bukan force kill.
func (s *Service) Kill(pid int) error {
	if IsProtected(pid) {
		return fmt.Errorf("PID %d adalah system process dan dilindungi", pid)
	}
	if !CanSignal(pid) {
		return fmt.Errorf("tidak punya permission untuk kill PID %d (milik user lain?)", pid)
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process %d tidak ditemukan: %w", pid, err)
	}
	if err := p.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("gagal mengirim SIGTERM ke PID %d: %w", pid, err)
	}
	return nil
}

// KillForce sends SIGKILL — hanya fallback dengan konfirmasi kedua (PRD §10, §23.5).
func (s *Service) KillForce(pid int) error {
	if IsProtected(pid) {
		return fmt.Errorf("PID %d adalah system process dan dilindungi", pid)
	}
	if !CanSignal(pid) {
		return fmt.Errorf("tidak punya permission untuk kill PID %d", pid)
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process %d tidak ditemukan: %w", pid, err)
	}
	if err := p.Signal(syscall.SIGKILL); err != nil {
		return fmt.Errorf("gagal mengirim SIGKILL ke PID %d: %w", pid, err)
	}
	return nil
}

// IsRoot reports whether running as root (PRD §23.1: jangan jalan sebagai root).
func IsRoot() bool { return os.Geteuid() == 0 }

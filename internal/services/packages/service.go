package packages

import (
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"lele-dev/internal/models"
	"lele-dev/internal/ui/shared"
)

// Manager keys (dipakai CLI --manager / positional + TUI filter).
const (
	ManagerNPM    = "npm"
	ManagerPip    = "pip"
	ManagerBrew   = "brew"
	ManagerPacman = "pacman"
	ManagerAUR    = "aur"
	ManagerApt    = "apt"
	ManagerCargo  = "cargo"
	ManagerGem    = "gem"
	ManagerGo     = "go"
)

// Order defines stable display order.
var Order = []string{
	ManagerNPM, ManagerPip, ManagerBrew, ManagerPacman, ManagerAUR,
	ManagerApt, ManagerCargo, ManagerGem, ManagerGo,
}

// Service lists third-party packages per package manager.
// Semua exec memakai fixed args (tanpa shell) — sesuai aturan security.
type Service struct{}

func New() *Service { return &Service{} }

// Managers reports supported managers and whether each is installed.
func (s *Service) Managers() []models.PackageManager {
	probes := []struct{ name, cmd string }{
		{ManagerNPM, "npm"},
		{ManagerPip, pipBin()},
		{ManagerBrew, "brew"},
		{ManagerPacman, "pacman"},
		{ManagerAUR, aurHelper()},
		{ManagerApt, "dpkg-query"},
		{ManagerCargo, "cargo"},
		{ManagerGem, "gem"},
		{ManagerGo, "go"},
	}
	out := make([]models.PackageManager, 0, len(probes))
	for _, p := range probes {
		bin := strings.Fields(p.cmd)[0]
		_, err := exec.LookPath(bin)
		out = append(out, models.PackageManager{
			Name:      p.name,
			Command:   p.cmd,
			Installed: err == nil,
		})
	}
	return out
}

// IsSupported reports whether name is a known manager key.
func IsSupported(name string) bool {
	for _, m := range Order {
		if m == strings.ToLower(name) {
			return true
		}
	}
	return false
}

// List returns packages for one manager, or all installed managers when empty.
// Deteksi tiap manager berjalan concurrent (PRD §39).
func (s *Service) List(manager string) ([]models.Package, error) {
	manager = strings.ToLower(strings.TrimSpace(manager))
	if manager != "" && !IsSupported(manager) {
		return nil, fmt.Errorf("package manager %q tidak dikenal (pilih: %s)", manager, strings.Join(Order, ", "))
	}
	targets := Order
	if manager != "" {
		targets = []string{manager}
	}
	var mu sync.Mutex
	var out []models.Package
	var wg sync.WaitGroup
	for _, m := range targets {
		wg.Add(1)
		go func(m string) {
			defer wg.Done()
			pkgs, err := s.listOne(m)
			if err != nil {
				return // manager tidak terinstall / gagal → skip, bukan fatal
			}
			mu.Lock()
			out = append(out, pkgs...)
			mu.Unlock()
		}(m)
	}
	wg.Wait()
	if manager != "" && len(out) == 0 {
		return nil, fmt.Errorf("%s tidak terinstall atau tidak ada package terdeteksi", manager)
	}
	attachSizes(out)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Manager != out[j].Manager {
			return orderIndex(out[i].Manager) < orderIndex(out[j].Manager)
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func orderIndex(m string) int {
	for i, k := range Order {
		if k == m {
			return i
		}
	}
	return len(Order)
}

func (s *Service) listOne(m string) ([]models.Package, error) {
	switch m {
	case ManagerNPM:
		return listNPM()
	case ManagerPip:
		return listPip()
	case ManagerBrew:
		return listBrew()
	case ManagerPacman:
		return listPacman()
	case ManagerAUR:
		return listAUR()
	case ManagerApt:
		return listApt()
	case ManagerCargo:
		return listCargo()
	case ManagerGem:
		return listGem()
	case ManagerGo:
		return listGoBins()
	}
	return nil, fmt.Errorf("unknown manager %q", m)
}

// ---- sizes ----

// attachSizes mengukur ukuran package secara concurrent.
// Pure-Go WalkDir (tanpa `du`) agar jalan di Linux & macOS.
func attachSizes(pkgs []models.Package) {
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup
	for i := range pkgs {
		if pkgs[i].Size != "" || pkgs[i].SizeDir == "" {
			continue // sudah ada ukuran (mis. database pacman/dpkg) atau tak terukur
		}
		wg.Add(1)
		go func(p *models.Package) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if n, err := dirSize(p.SizeDir); err == nil && n > 0 {
				p.SizeBytes = n
				p.Size = shared.FormatBytes(n)
			} else {
				p.Size = "—"
			}
		}(&pkgs[i])
	}
	wg.Wait()
	for i := range pkgs {
		if pkgs[i].Size == "" {
			pkgs[i].Size = "—"
		}
	}
}

// dirSize sums file sizes under dir (tidak mengikuti symlink).
func dirSize(dir string) (int64, error) {
	var total int64
	err := filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // file hilang di tengah jalan → skip
		}
		if d.Type().IsRegular() {
			if info, err := d.Info(); err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total, err
}

// Search filters packages by manager/name/version (TUI `/`).
func Search(all []models.Package, query string) []models.Package {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return all
	}
	var out []models.Package
	for _, p := range all {
		if strings.Contains(strings.ToLower(p.Manager), q) ||
			strings.Contains(strings.ToLower(p.Name), q) ||
			strings.Contains(strings.ToLower(p.Version), q) {
			out = append(out, p)
		}
	}
	return out
}

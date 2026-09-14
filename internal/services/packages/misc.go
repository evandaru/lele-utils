package packages

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"lele-dev/internal/models"
	"lele-dev/internal/ui/shared"
)

// ---- cargo ----

// listCargo parses `cargo install --list`; size = jumlah binary terinstall.
func listCargo() ([]models.Package, error) {
	if _, err := exec.LookPath("cargo"); err != nil {
		return nil, err
	}
	out, err := exec.Command("cargo", "install", "--list").Output()
	if err != nil && len(out) == 0 {
		return nil, err
	}
	entries := ParseCargoList(string(out))
	binDir := cargoBinDir()
	var pkgs []models.Package
	for _, e := range entries {
		p := models.Package{Manager: ManagerCargo, Name: e.name, Version: e.version}
		var total int64
		for _, b := range e.bins {
			if st, err := os.Stat(filepath.Join(binDir, b)); err == nil && !st.IsDir() {
				total += st.Size()
			}
		}
		if total > 0 {
			p.SizeBytes = total
			p.Size = shared.FormatBytes(total)
		}
		pkgs = append(pkgs, p)
	}
	return pkgs, nil
}

type cargoEntry struct {
	name    string
	version string
	bins    []string
}

// ParseCargoList parses `cargo install --list`. Pure → tested.
func ParseCargoList(out string) []cargoEntry {
	var entries []cargoEntry
	header := regexp.MustCompile(`^(\S+)\s+v(\S+):$`)
	for _, line := range strings.Split(out, "\n") {
		if m := header.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			entries = append(entries, cargoEntry{name: m[1], version: m[2]})
		} else if bin := strings.TrimSpace(line); bin != "" && len(entries) > 0 && !strings.Contains(bin, " ") {
			entries[len(entries)-1].bins = append(entries[len(entries)-1].bins, bin)
		}
	}
	return entries
}

func cargoBinDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return filepath.Join(h, ".cargo", "bin")
	}
	return ""
}

// ---- gem ----

// listGem parses `gem list`; size = direktori gem versi pertama.
func listGem() ([]models.Package, error) {
	if _, err := exec.LookPath("gem"); err != nil {
		return nil, err
	}
	out, err := exec.Command("gem", "list").Output()
	if err != nil && len(out) == 0 {
		return nil, err
	}
	entries := ParseGemList(string(out))
	gemDir := gemHome()
	var pkgs []models.Package
	for _, e := range entries {
		p := models.Package{Manager: ManagerGem, Name: e.name, Version: e.version}
		if gemDir != "" && e.firstVer != "" {
			p.SizeDir = filepath.Join(gemDir, "gems", e.name+"-"+e.firstVer)
		}
		pkgs = append(pkgs, p)
	}
	return pkgs, nil
}

type gemEntry struct {
	name     string
	version  string // semua versi, dipisah ", "
	firstVer string
}

// ParseGemList parses `gem list` ("name (v1, v2)" / "(default: v)"). Pure → tested.
func ParseGemList(out string) []gemEntry {
	var entries []gemEntry
	re := regexp.MustCompile(`^(\S+)\s+\(([^)]+)\)$`)
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		m := re.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		raw := strings.Split(m[2], ",")
		var vers []string
		for _, v := range raw {
			v = strings.TrimSpace(v)
			v = strings.TrimPrefix(v, "default: ")
			if v != "" {
				vers = append(vers, v)
			}
		}
		if len(vers) == 0 {
			continue
		}
		entries = append(entries, gemEntry{name: m[1], version: strings.Join(vers, ", "), firstVer: vers[0]})
	}
	return entries
}

func gemHome() string {
	out, err := exec.Command("gem", "environment", "gemdir").Output()
	if err != nil {
		return ""
	}
	return trimLine(string(out))
}

// ---- go binaries ----

// listGoBins lists executables in GOPATH/bin (version tak terlacak → "—").
func listGoBins() ([]models.Package, error) {
	if _, err := exec.LookPath("go"); err != nil {
		return nil, err
	}
	out, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return nil, err
	}
	binDir := filepath.Join(trimLine(string(out)), "bin")
	entries, err := os.ReadDir(binDir)
	if err != nil {
		return nil, err
	}
	var files []GoBin
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		var sz int64
		if info, err := e.Info(); err == nil {
			sz = info.Size()
		}
		files = append(files, GoBin{Name: e.Name(), Size: sz})
	}
	return ParseGoBins(files), nil
}

// GoBin is one executable file (disuntik agar testable).
type GoBin struct {
	Name string
	Size int64
}

// ParseGoBins converts bin files to packages (versi tak terlacak → "—"). Pure → tested.
func ParseGoBins(files []GoBin) []models.Package {
	pkgs := make([]models.Package, 0, len(files))
	for _, f := range files {
		p := models.Package{Manager: ManagerGo, Name: f.Name, Version: "—"}
		if f.Size > 0 {
			p.SizeBytes = f.Size
			p.Size = shared.FormatBytes(f.Size)
		}
		pkgs = append(pkgs, p)
	}
	return pkgs
}

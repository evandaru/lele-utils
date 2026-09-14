package packages

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"

	"lele-dev/internal/models"
)

// pipBin picks the available pip invocation.
func pipBin() string {
	if _, err := exec.LookPath("pip"); err == nil {
		return "pip"
	}
	if _, err := exec.LookPath("pip3"); err == nil {
		return "pip3"
	}
	return "pip"
}

func pipArgs(args ...string) *exec.Cmd {
	bin := pipBin()
	if bin == "pip" {
		if _, err := exec.LookPath("pip"); err != nil {
			if _, err2 := exec.LookPath("pip3"); err2 == nil {
				return exec.Command("pip3", args...)
			}
			return exec.Command("python3", append([]string{"-m", "pip"}, args...)...)
		}
	}
	return exec.Command(bin, args...)
}

// listPip returns `pip list` packages + install locations for sizing.
func listPip() ([]models.Package, error) {
	out, err := pipArgs("list", "--format=json").Output()
	if err != nil && len(out) == 0 {
		return nil, err
	}
	pkgs := ParsePipListJSON(out)
	if len(pkgs) == 0 {
		return pkgs, nil
	}
	// Satu panggilan `pip show` untuk semua lokasi (hemat fork).
	names := make([]string, 0, len(pkgs))
	for _, p := range pkgs {
		names = append(names, p.Name)
	}
	locs := map[string]string{}
	if show, err := pipArgs(append([]string{"show"}, names...)...).Output(); err == nil {
		locs = ParsePipShow(string(show))
	}
	for i := range pkgs {
		pkgs[i].Manager = ManagerPip
		if loc, ok := locs[strings.ToLower(pkgs[i].Name)]; ok {
			pkgs[i].SizeDir = pipGuessDir(loc, pkgs[i].Name)
		}
	}
	return pkgs, nil
}

// ParsePipListJSON parses `pip list --format=json`. Pure → tested.
func ParsePipListJSON(data []byte) []models.Package {
	var v []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if json.Unmarshal(data, &v) != nil {
		return nil
	}
	out := make([]models.Package, 0, len(v))
	for _, p := range v {
		ver := p.Version
		if ver == "" {
			ver = "—"
		}
		out = append(out, models.Package{Name: p.Name, Version: ver})
	}
	return out
}

// ParsePipShow parses `pip show ...` blocks into lower(name) → Location. Pure → tested.
func ParsePipShow(out string) map[string]string {
	locs := map[string]string{}
	var name string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "Name: ") {
			name = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "Name: ")))
		} else if strings.HasPrefix(line, "Location: ") && name != "" {
			locs[name] = strings.TrimSpace(strings.TrimPrefix(line, "Location: "))
			name = ""
		} else if strings.TrimSpace(line) == "---" {
			name = ""
		}
	}
	return locs
}

// pipGuessDir menebak direktori package (tolerant terhadap -/_/case).
func pipGuessDir(loc, name string) string {
	candidates := []string{
		name,
		strings.ReplaceAll(name, "-", "_"),
		strings.ReplaceAll(name, "_", "-"),
		strings.ToLower(name),
		strings.ToLower(strings.ReplaceAll(name, "-", "_")),
	}
	for _, c := range candidates {
		if dirExists(filepath.Join(loc, c)) {
			return filepath.Join(loc, c)
		}
	}
	return ""
}

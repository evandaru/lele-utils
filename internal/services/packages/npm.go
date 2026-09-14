package packages

import (
	"encoding/json"
	"os/exec"
	"path/filepath"

	"lele-dev/internal/models"
)

// listNPM returns global npm packages via `npm ls -g --depth=0 --json`.
func listNPM() ([]models.Package, error) {
	if _, err := exec.LookPath("npm"); err != nil {
		return nil, err
	}
	out, err := exec.Command("npm", "ls", "-g", "--depth=0", "--json").Output()
	if err != nil && len(out) == 0 {
		return nil, err
	}
	pkgs := ParseNpmLsJSON(out)
	root := npmRoot()
	for i := range pkgs {
		pkgs[i].Manager = ManagerNPM
		if root != "" {
			pkgs[i].SizeDir = filepath.Join(root, pkgs[i].Name)
		}
	}
	return pkgs, nil
}

func npmRoot() string {
	out, err := exec.Command("npm", "root", "-g").Output()
	if err != nil {
		return ""
	}
	return trimLine(string(out))
}

// ParseNpmLsJSON parses `npm ls -g --depth=0 --json`. Pure → tested.
func ParseNpmLsJSON(data []byte) []models.Package {
	var v struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}
	if json.Unmarshal(data, &v) != nil {
		return nil
	}
	var out []models.Package
	for name, d := range v.Dependencies {
		ver := d.Version
		if ver == "" {
			ver = "—"
		}
		out = append(out, models.Package{Name: name, Version: ver})
	}
	return out
}

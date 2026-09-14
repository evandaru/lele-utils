package packages

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"lele-dev/internal/models"
)

// listBrew returns `brew list --versions` + cellar dirs for sizing.
func listBrew() ([]models.Package, error) {
	if _, err := exec.LookPath("brew"); err != nil {
		return nil, err
	}
	out, err := exec.Command("brew", "list", "--versions").Output()
	if err != nil {
		return nil, err
	}
	pkgs := ParseBrewVersions(string(out))
	cellar := brewCellar()
	for i := range pkgs {
		pkgs[i].Manager = ManagerBrew
		if cellar != "" {
			pkgs[i].SizeDir = filepath.Join(cellar, pkgs[i].Name)
		}
	}
	return pkgs, nil
}

func brewCellar() string {
	out, err := exec.Command("brew", "--cellar").Output()
	if err != nil {
		return ""
	}
	return trimLine(string(out))
}

// ParseBrewVersions parses `brew list --versions` ("name v1 v2..."). Pure → tested.
func ParseBrewVersions(out string) []models.Package {
	var pkgs []models.Package
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ver := fields[1]
		if len(fields) > 2 {
			ver += " (+" + strconv.Itoa(len(fields)-2) + ")"
		}
		pkgs = append(pkgs, models.Package{Name: fields[0], Version: ver})
	}
	return pkgs
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func trimLine(s string) string {
	return strings.TrimSpace(strings.SplitN(s, "\n", 2)[0])
}

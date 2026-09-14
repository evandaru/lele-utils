package packages

import (
	"os/exec"
	"strconv"
	"strings"

	"lele-dev/internal/models"
	"lele-dev/internal/ui/shared"
)

// listApt returns installed deb packages via single dpkg-query call
// (name, version, Installed-Size dalam KB).
func listApt() ([]models.Package, error) {
	if _, err := exec.LookPath("dpkg-query"); err != nil {
		return nil, err
	}
	out, err := exec.Command("dpkg-query", "-W", "-f=${Package}\t${Version}\t${Installed-Size}\n").Output()
	if err != nil {
		return nil, err
	}
	return ParseDpkgQuery(string(out)), nil
}

// ParseDpkgQuery parses dpkg-query -W tabular output. Pure → tested.
func ParseDpkgQuery(out string) []models.Package {
	var pkgs []models.Package
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		name, ver := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if name == "" {
			continue
		}
		if ver == "" {
			ver = "—"
		}
		p := models.Package{Manager: ManagerApt, Name: name, Version: ver}
		if kb, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64); err == nil && kb > 0 {
			p.SizeBytes = kb * 1024
			p.Size = shared.FormatBytes(p.SizeBytes)
		}
		pkgs = append(pkgs, p)
	}
	return pkgs
}

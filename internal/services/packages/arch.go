package packages

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"lele-dev/internal/models"
	"lele-dev/internal/ui/shared"
)

// aurHelper picks yay → paru → pacman (fallback) for foreign packages.
func aurHelper() string {
	for _, b := range []string{"yay", "paru", "pacman"} {
		if _, err := exec.LookPath(b); err == nil {
			return b
		}
	}
	return "yay"
}

// listPacman returns repo-native packages (`pacman -Qn`) + sizes from `pacman -Qi`.
func listPacman() ([]models.Package, error) {
	if _, err := exec.LookPath("pacman"); err != nil {
		return nil, err
	}
	out, err := exec.Command("pacman", "-Qn").Output()
	if err != nil {
		return nil, err
	}
	pkgs := ParsePacmanQ(string(out), ManagerPacman)
	applyPacmanSizes(pkgs)
	return pkgs, nil
}

// listAUR returns foreign packages (yay/paru -Qm, fallback pacman -Qm).
func listAUR() ([]models.Package, error) {
	helper := aurHelper()
	if _, err := exec.LookPath(helper); err != nil {
		return nil, err
	}
	out, err := exec.Command(helper, "-Qm").Output()
	if err != nil && len(out) == 0 {
		return nil, err
	}
	pkgs := ParsePacmanQ(string(out), ManagerAUR)
	applyPacmanSizes(pkgs)
	return pkgs, nil
}

// applyPacmanSizes mengisi Size dari SATU panggilan `pacman -Qi` (efisien).
func applyPacmanSizes(pkgs []models.Package) {
	out, err := exec.Command("pacman", "-Qi").Output()
	if err != nil {
		return
	}
	sizes := ParsePacmanQi(string(out))
	for i := range pkgs {
		if n, ok := sizes[pkgs[i].Name]; ok && n > 0 {
			pkgs[i].SizeBytes = n
			pkgs[i].Size = shared.FormatBytes(n)
		}
	}
}

// ParsePacmanQ parses `pacman -Q[n|m]` ("name version" per baris). Pure → tested.
func ParsePacmanQ(out, manager string) []models.Package {
	var pkgs []models.Package
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name, ver, _ := strings.Cut(line, " ")
		ver = strings.TrimSpace(ver)
		if name == "" || ver == "" {
			continue
		}
		pkgs = append(pkgs, models.Package{Manager: manager, Name: name, Version: ver})
	}
	return pkgs
}

// ParsePacmanQi parses `pacman -Qi` menjadi name → installed bytes. Pure → tested.
func ParsePacmanQi(out string) map[string]int64 {
	sizes := map[string]int64{}
	var name string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "Name") {
			name = fieldValue(line)
		} else if strings.HasPrefix(line, "Installed Size") && name != "" {
			if n := ParsePacmanSize(fieldValue(line)); n > 0 {
				sizes[name] = n
			}
			name = ""
		} else if strings.TrimSpace(line) == "" {
			name = ""
		}
	}
	return sizes
}

func fieldValue(line string) string {
	_, v, _ := strings.Cut(line, ":")
	return strings.TrimSpace(v)
}

// ParsePacmanSize converts "12.30 MiB" / "225.00 KiB" / "512 B" → bytes. Pure → tested.
func ParsePacmanSize(s string) int64 {
	m := regexp.MustCompile(`([\d.]+)\s*([KMGT]?i?B)`).FindStringSubmatch(s)
	if m == nil {
		return 0
	}
	f, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0
	}
	mult := 1.0
	switch m[2] {
	case "KiB", "KB":
		mult = 1024
	case "MiB", "MB":
		mult = 1024 * 1024
	case "GiB", "GB":
		mult = 1024 * 1024 * 1024
	case "TiB", "TB":
		mult = 1024 * 1024 * 1024 * 1024
	}
	return int64(f * mult)
}

package tools

import (
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"dev-utils/internal/models"
)

type spec struct {
	Name    string
	Command string
	Args    []string
}

func specs() []spec {
	return []spec{
		{"Git", "git", []string{"--version"}},
		{"Docker", "docker", []string{"--version"}},
		{"Podman", "podman", []string{"--version"}},
		{"Make", "make", []string{"--version"}},
		{"CMake", "cmake", []string{"--version"}},
		{"NPM", "npm", []string{"--version"}},
		{"PNPM", "pnpm", []string{"--version"}},
		{"Yarn", "yarn", []string{"--version"}},
		{"Composer", "composer", []string{"--version"}},
		{"Cargo", "cargo", []string{"--version"}},
		{"ADB", "adb", []string{"--version"}},
		{"Gradle", "gradle", []string{"--version"}},
		{"Go", "go", []string{"version"}},
		{"Node.js", "node", []string{"--version"}},
		{"Python", "python3", []string{"--version"}},
	}
}

var versionRe = regexp.MustCompile(`\d+\.\d+(\.\d+)?`)

// Service — PRD §14.
type Service struct{}

func New() *Service { return &Service{} }

func (s *Service) DetectAll() []models.Tool {
	sp := specs()
	out := make([]models.Tool, len(sp))
	var wg sync.WaitGroup
	for i, t := range sp {
		wg.Add(1)
		go func(i int, t spec) {
			defer wg.Done()
			out[i] = detectOne(t)
		}(i, t)
	}
	wg.Wait()
	return out
}

func detectOne(s spec) models.Tool {
	t := models.Tool{Name: s.Name, Command: s.Command}
	path, err := exec.LookPath(s.Command)
	if err != nil {
		return t
	}
	t.Path = path
	t.Installed = true
	if b, err := exec.Command(s.Command, s.Args...).CombinedOutput(); err == nil || len(b) > 0 {
		t.Version = versionRe.FindString(string(b))
		if t.Version == "" {
			t.Version = strings.TrimSpace(strings.Split(string(b), "\n")[0])
			if len(t.Version) > 40 {
				t.Version = t.Version[:40]
			}
		}
	}
	return t
}

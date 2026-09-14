package container

import (
	"os/exec"
	"strconv"
	"strings"

	"lele-dev/internal/models"
)

// Service — PRD §15. Support Docker + Podman, jangan anggap Docker harus ada.
type Service struct{}

func New() *Service { return &Service{} }

// Available reports which runtimes exist.
func Available() (docker, podman bool) {
	// exec.LookPath returns (string, error); ada = err==nil
	docker = hasBin("docker")
	podman = hasBin("podman")
	return docker, podman
}

func hasBin(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// List queries docker and podman (yang tersedia saja).
func (s *Service) List() ([]models.Container, error) {
	var out []models.Container
	for _, rt := range []string{"docker", "podman"} {
		if !hasBin(rt) {
			continue
		}
		cs, err := listRuntime(rt)
		if err != nil {
			continue // daemon mati → skip, bukan fatal
		}
		out = append(out, cs...)
	}
	return out, nil
}

func listRuntime(rt string) ([]models.Container, error) {
	// Format tab-separated agar mudah parse. Tanpa shell — fixed args (PRD §23.6).
	out, err := exec.Command(rt, "ps", "--format",
		"{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.State}}\t{{.Ports}}").Output()
	if err != nil {
		return nil, err
	}
	return ParsePsOutput(rt, string(out)), nil
}

// ParsePsOutput parses docker/podman ps --format output. Pure → tested.
func ParsePsOutput(runtime, out string) []models.Container {
	var cs []models.Container
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		for len(parts) < 6 {
			parts = append(parts, "")
		}
		cs = append(cs, models.Container{
			ID:      parts[0],
			Name:    parts[1],
			Image:   parts[2],
			Status:  parts[3],
			State:   parts[4],
			Ports:   parts[5],
			Runtime: runtime,
		})
	}
	return cs
}

// Start/Stop/Restart/Logs/Kill — fixed args, tanpa arbitrary shell (PRD §23.6).
func (s *Service) Start(rt, id string) error {
	return exec.Command(rt, "start", id).Run()
}

func (s *Service) Stop(rt, id string) error {
	return exec.Command(rt, "stop", id).Run()
}

func (s *Service) Restart(rt, id string) error {
	return exec.Command(rt, "restart", id).Run()
}

func (s *Service) Kill(rt, id string) error {
	return exec.Command(rt, "kill", id).Run()
}

func (s *Service) Logs(rt, id string, tail int) (string, error) {
	if tail <= 0 {
		tail = 50
	}
	b, err := exec.Command(rt, "logs", "--tail", strconv.Itoa(tail), id).CombinedOutput()
	return string(b), err
}

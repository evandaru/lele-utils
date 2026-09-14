package doctor

import (
	"os/exec"

	"dev-utils/internal/models"
	"dev-utils/internal/services/runtime"
	"dev-utils/internal/services/tools"
)

// Service — PRD §20 + §36. Diagnostic developer environment.
type Service struct {
	runtimes *runtime.Service
	toolsSvc *tools.Service
}

func New() *Service {
	return &Service{runtimes: runtime.New(), toolsSvc: tools.New()}
}

// Check runs full diagnostics concurrently.
func (s *Service) Check() *models.DoctorReport {
	report := &models.DoctorReport{}
	rts := s.runtimes.DetectAll()
	for _, r := range rts {
		e := models.DoctorEntry{Name: r.Name, Version: r.Version}
		switch {
		case r.Installed && r.Version != "":
			e.Status = "ok"
			e.Detail = "installed (" + r.Path + ")"
		case r.Installed:
			e.Status = "warning"
			e.Detail = "installed tapi versi tidak terdeteksi"
		default:
			e.Status = "missing"
			e.Detail = "not installed"
		}
		report.Entries = append(report.Entries, e)
	}
	tls := s.toolsSvc.DetectAll()
	seen := map[string]bool{}
	for _, r := range rts {
		seen[r.Name] = true
	}
	for _, t := range tls {
		if seen[t.Name] {
			continue
		}
		e := models.DoctorEntry{Name: t.Name, Version: t.Version}
		switch {
		case t.Installed && t.Version != "":
			e.Status = "ok"
			e.Detail = "installed (" + t.Path + ")"
		case t.Installed:
			e.Status = "warning"
			e.Detail = "installed tapi versi tidak terdeteksi"
		default:
			e.Status = "missing"
			e.Detail = "not installed"
		}
		report.Entries = append(report.Entries, e)
	}
	// Cek khusus: docker daemon.
	report.Entries = append(report.Entries, dockerDaemonEntry())
	// Hitung ringkasan.
	for _, e := range report.Entries {
		switch e.Status {
		case "ok":
			report.OK++
		case "warning":
			report.Warnings++
		default:
			report.Missing++
		}
	}
	return report
}

func dockerDaemonEntry() models.DoctorEntry {
	if _, err := exec.LookPath("docker"); err != nil {
		return models.DoctorEntry{Name: "Docker daemon", Status: "missing", Detail: "docker CLI tidak terinstall"}
	}
	if err := exec.Command("docker", "info").Run(); err != nil {
		return models.DoctorEntry{Name: "Docker daemon", Status: "warning", Detail: "installed, daemon not running"}
	}
	return models.DoctorEntry{Name: "Docker daemon", Status: "ok", Detail: "running"}
}

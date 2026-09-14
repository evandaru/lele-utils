package platform

import (
	"lele-dev/internal/models"
)

// ProcessService — PRD §24. UI tidak boleh tahu perbedaan OS.
type ProcessService interface {
	List() ([]models.Process, error)
	Kill(pid int) error      // SIGTERM (default)
	KillForce(pid int) error // SIGKILL (fallback, konfirmasi kedua)
	Find(pid int) (*models.Process, error)
}

// PortService maps listening ports → PID → process.
type PortService interface {
	List() ([]models.Port, error)
	Find(port int) (*models.Port, error)
}

// NetworkService eig. interfaces, gateway, DNS tanpa asumsi nama.
type NetworkService interface {
	Info() (*models.NetworkInfo, error)
}

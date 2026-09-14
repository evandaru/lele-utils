package network

import (
	"lele-dev/internal/models"
	"lele-dev/internal/platform/darwin"
	"lele-dev/internal/platform/linux"
	"runtime"
)

// Service — PRD §13. Tanpa asumsi nama interface.
type Service struct{}

func New() *Service { return &Service{} }

func (s *Service) Info() (*models.NetworkInfo, error) {
	if runtime.GOOS == "darwin" {
		return darwin.NetworkInfo()
	}
	return linux.NetworkInfo()
}

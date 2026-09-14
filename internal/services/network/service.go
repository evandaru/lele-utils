package network

import (
	"dev-utils/internal/models"
	"dev-utils/internal/platform/darwin"
	"dev-utils/internal/platform/linux"
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

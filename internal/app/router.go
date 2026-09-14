package app

import (
	"dev-utils/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the TUI workstation (PRD §9).
func Run(cfg config.Config) error {
	m := New(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

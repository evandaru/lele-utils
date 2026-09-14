package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

// Config — PRD §22. Semua field optional dengan default wajar.
type Config struct {
	RefreshInterval int    `toml:"refresh_interval"`
	DefaultScreen   string `toml:"default_screen"`
	ConfirmKill     bool   `toml:"confirm_kill"`
	ShowSystem      bool   `toml:"show_system"`
}

// Default returns sane defaults. Aplikasi harus jalan tanpa config file.
func Default() Config {
	return Config{
		RefreshInterval: 3,
		DefaultScreen:   "dashboard",
		ConfirmKill:     true,
		ShowSystem:      true,
	}
}

// Path returns platform-specific config path.
// Linux: ~/.config/lele-dev/config.toml
// macOS: ~/Library/Application Support/lele-dev/config.toml
func Path() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "lele-dev", "config.toml")
	}
	// Linux + fallback: XDG_CONFIG_HOME atau ~/.config
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "lele-dev", "config.toml")
	}
	return filepath.Join(home, ".config", "lele-dev", "config.toml")
}

// Load reads config file if present, else returns defaults.
// Tidak pernah error fatal: file hilang = defaults.
func Load() Config {
	cfg := Default()
	p := Path()
	if p == "" {
		return cfg
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return cfg // file tidak ada → defaults
	}
	var file Config
	// Decode dengan cara yang toleran terhadap field kosong:
	// mulai dari defaults lalu timpa yang ada di file.
	file = cfg
	if _, err := toml.Decode(string(data), &file); err != nil {
		return cfg
	}
	if file.RefreshInterval <= 0 {
		file.RefreshInterval = cfg.RefreshInterval
	}
	if file.DefaultScreen == "" {
		file.DefaultScreen = cfg.DefaultScreen
	}
	return file
}

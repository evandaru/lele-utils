package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultDoesNotRequireFile(t *testing.T) {
	cfg := Default()
	if cfg.RefreshInterval != 3 || cfg.DefaultScreen != "dashboard" || !cfg.ConfirmKill || !cfg.ShowSystem {
		t.Errorf("unexpected defaults: %+v", cfg)
	}
}

func TestPathPlatformSpecific(t *testing.T) {
	p := Path()
	if p == "" {
		t.Fatal("empty config path")
	}
	if !filepath.IsAbs(p) {
		t.Errorf("config path not absolute: %q", p)
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	// Pindahkan HOME sementara agar file tidak ditemukan.
	old := os.Getenv("HOME")
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	_ = old
	_ = oldXDG
	cfg := Load()
	if cfg.RefreshInterval != 3 {
		t.Errorf("got %+v, want defaults", cfg)
	}
}

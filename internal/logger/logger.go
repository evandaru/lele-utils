package logger

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	mu      sync.Mutex
	enabled = false
	logger  *log.Logger
)

// Init mengaktifkan file logging bila --debug diberikan (PRD §28).
// Default: silent. Jangan campur debug log dengan TUI.
func Init(debug bool) {
	mu.Lock()
	defer mu.Unlock()
	enabled = debug
	if !debug {
		return
	}
	p := FilePath()
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	logger = log.New(f, "dev-utils ", log.LstdFlags|log.Lshortfile)
}

// FilePath returns platform-appropriate state dir.
func FilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "dev-utils.log")
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "dev-utils", "dev-utils.log")
	}
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, "dev-utils", "dev-utils.log")
	}
	return filepath.Join(home, ".local", "state", "dev-utils", "dev-utils.log")
}

// Debugf logs only when debug enabled.
func Debugf(format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	if enabled && logger != nil {
		logger.Printf(format, args...)
	}
}

// Enabled reports debug state.
func Enabled() bool { return enabled }

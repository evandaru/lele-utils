package runtime

import (
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"lele-dev/internal/models"
)

// Detector — PRD §12.
type Detector interface {
	Detect() (*models.Runtime, error)
}

type spec struct {
	Name    string
	Command string
	Args    []string
	Parse   func(string) string
}

// Semua runtime yang didukung PRD §12.
func specs() []spec {
	return []spec{
		{"Go", "go", []string{"version"}, ParseGoVersion},
		{"Node.js", "node", []string{"--version"}, ParseNodeVersion},
		{"Bun", "bun", []string{"--version"}, ParseGenericVersion},
		{"Deno", "deno", []string{"--version"}, ParseDenoVersion},
		{"Python", "python3", []string{"--version"}, ParsePythonVersion},
		{"PHP", "php", []string{"--version"}, ParsePHPVersion},
		{"Ruby", "ruby", []string{"--version"}, ParseRubyVersion},
		{"Rust", "rustc", []string{"--version"}, ParseRustVersion},
		{"Java", "java", []string{"--version"}, ParseJavaVersion},
		{"Kotlin", "kotlin", []string{"-version"}, ParseKotlinVersion},
		{"Dart", "dart", []string{"--version"}, ParseDartVersion},
		{"Flutter", "flutter", []string{"--version"}, ParseFlutterVersion},
		{"Swift", "swift", []string{"--version"}, ParseSwiftVersion},
		{"GCC", "gcc", []string{"--version"}, ParseGCCVersion},
		{"Clang", "clang", []string{"--version"}, ParseClangVersion},
	}
}

// Service detects runtimes concurrently (PRD §39).
type Service struct{}

func New() *Service { return &Service{} }

// DetectAll runs all detectors concurrently, mengembalikan yang installed + yang tidak.
func (s *Service) DetectAll() []models.Runtime {
	sp := specs()
	out := make([]models.Runtime, len(sp))
	var wg sync.WaitGroup
	for i, s := range sp {
		wg.Add(1)
		go func(i int, s spec) {
			defer wg.Done()
			out[i] = detectOne(s)
		}(i, s)
	}
	wg.Wait()
	return out
}

// InstalledOnly returns only detected runtimes.
func (s *Service) InstalledOnly() []models.Runtime {
	var out []models.Runtime
	for _, r := range s.DetectAll() {
		if r.Installed {
			out = append(out, r)
		}
	}
	return out
}

func detectOne(s spec) models.Runtime {
	rt := models.Runtime{Name: s.Name, Command: s.Command}
	// PRD §12: gunakan exec.LookPath, bukan hanya env var.
	path, err := exec.LookPath(s.Command)
	if err != nil {
		return rt // Installed=false
	}
	rt.Path = path
	// Dapatkan versi sebenarnya via exec.
	cmd := exec.Command(s.Command, s.Args...)
	b, err := cmd.CombinedOutput()
	if err != nil && len(b) == 0 {
		rt.Installed = true
		return rt
	}
	rt.Version = s.Parse(strings.TrimSpace(string(b)))
	rt.Installed = true
	return rt
}

// ---- Version parsers (pure, unit tested) ----

var versionRe = regexp.MustCompile(`\d+\.\d+(\.\d+)?`)

func firstVersion(s string) string {
	return versionRe.FindString(s)
}

// ParseGoVersion: "go version go1.25.1 linux/amd64" → "1.25.1".
func ParseGoVersion(s string) string {
	if m := regexp.MustCompile(`go(\d+\.\d+(\.\d+)?)`).FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return firstVersion(s)
}

// ParseNodeVersion: "v24.8.0" → "24.8.0".
func ParseNodeVersion(s string) string { return strings.TrimPrefix(firstVersion(s), "v") }

// ParseGenericVersion: "1.2.21" → "1.2.21".
func ParseGenericVersion(s string) string { return firstVersion(s) }

// ParseDenoVersion: "deno 2.x ..." → versi pertama.
func ParseDenoVersion(s string) string { return firstVersion(s) }

// ParsePythonVersion: "Python 3.13.7" → "3.13.7".
func ParsePythonVersion(s string) string { return firstVersion(s) }

// ParsePHPVersion: "PHP 8.4.15 (cli) ..." → "8.4.15".
func ParsePHPVersion(s string) string { return firstVersion(s) }

// ParseRubyVersion: "ruby 3.4.x" → versi.
func ParseRubyVersion(s string) string { return firstVersion(s) }

// ParseRustVersion: "rustc 1.89.0 ..." → "1.89.0".
func ParseRustVersion(s string) string { return firstVersion(s) }

// ParseJavaVersion: 'openjdk 25 2025-09-16' → "25", atau 'openjdk version "17.0.x"'.
func ParseJavaVersion(s string) string {
	if m := regexp.MustCompile(`"(\d+[\.\d_]*)"`).FindStringSubmatch(s); m != nil {
		return strings.ReplaceAll(m[1], "_", ".")
	}
	// Java 9+ mencetak major tunggal: "openjdk 25 ...", "java 25 ...".
	if m := regexp.MustCompile(`(?:openjdk|java)\s+(\d+[\.\d_]*)`).FindStringSubmatch(s); m != nil {
		return strings.ReplaceAll(m[1], "_", ".")
	}
	return firstVersion(s)
}

// ParseKotlinVersion: "Kotlin version 2.x".
func ParseKotlinVersion(s string) string { return firstVersion(s) }

// ParseDartVersion: "Dart SDK version: 3.x".
func ParseDartVersion(s string) string { return firstVersion(s) }

// ParseFlutterVersion: "Flutter 3.x • ...".
func ParseFlutterVersion(s string) string { return firstVersion(s) }

// ParseSwiftVersion: "Swift version 6.x".
func ParseSwiftVersion(s string) string { return firstVersion(s) }

// ParseGCCVersion: "gcc (GCC) 14.x".
func ParseGCCVersion(s string) string { return firstVersion(s) }

// ParseClangVersion: "clang version 18.x".
func ParseClangVersion(s string) string { return firstVersion(s) }

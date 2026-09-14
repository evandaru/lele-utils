package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"dev-utils/internal/models"
)

// Service — PRD §16. Deteksi project dari file penanda.
type Service struct{}

func New() *Service { return &Service{} }

// Inspect analyzes dir (default: cwd).
func (s *Service) Inspect(dir string) (*models.ProjectInfo, error) {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	abs, _ := filepath.Abs(dir)
	info := &models.ProjectInfo{
		Name: filepath.Base(abs),
		Path: abs,
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}
	files := map[string]bool{}
	for _, e := range entries {
		if !e.IsDir() {
			files[e.Name()] = true
		}
	}
	addFile := func(n string) {
		if files[n] {
			info.Files = append(info.Files, n)
		}
	}
	for _, n := range []string{"go.mod", "package.json", "composer.json", "Cargo.toml",
		"requirements.txt", "pyproject.toml", "Gemfile", "pom.xml", "build.gradle",
		"Dockerfile", "compose.yaml", "docker-compose.yml", ".env", "README.md",
		"pubspec.yaml", "Package.swift"} {
		addFile(n)
	}

	has := func(n string) bool { return files[n] }

	if has("go.mod") {
		info.Languages = append(info.Languages, "Go")
		info.Tools = append(info.Tools, "Go")
	}
	if has("package.json") {
		info.Languages = append(info.Languages, "Node.js")
		info.PackageManager = detectNodePM(abs, files)
		info.Framework = detectNodeFramework(abs)
		info.Tools = append(info.Tools, "Node.js")
	}
	if has("composer.json") {
		info.Languages = append(info.Languages, "PHP")
		info.Tools = append(info.Tools, "PHP", "Composer")
	}
	if has("Cargo.toml") {
		info.Languages = append(info.Languages, "Rust")
		info.Tools = append(info.Tools, "Cargo")
	}
	if has("requirements.txt") || has("pyproject.toml") {
		info.Languages = append(info.Languages, "Python")
	}
	if has("Gemfile") {
		info.Languages = append(info.Languages, "Ruby")
	}
	if has("pubspec.yaml") {
		info.Languages = append(info.Languages, "Dart/Flutter")
		info.Tools = append(info.Tools, "Flutter")
	}
	if has("Package.swift") {
		info.Languages = append(info.Languages, "Swift")
	}
	if has("Dockerfile") || has("compose.yaml") || has("docker-compose.yml") {
		info.Tools = append(info.Tools, "Docker")
	}
	if has(".env") {
		info.Tools = append(info.Tools, "env")
	}
	if branch := gitBranch(abs); branch != "" {
		info.GitBranch = branch
	}
	return info, nil
}

func detectNodePM(dir string, files map[string]bool) string {
	switch {
	case files["pnpm-lock.yaml"]:
		return "pnpm"
	case files["yarn.lock"]:
		return "yarn"
	case files["bun.lockb"], files["bun.lock"]:
		return "bun"
	case files["deno.json"], files["deno.jsonc"]:
		return "deno"
	default:
		if files["package-lock.json"] {
			return "npm"
		}
		// cek packageManager field
		if data, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
			var pj struct {
				PackageManager string `json:"packageManager"`
			}
			if json.Unmarshal(data, &pj) == nil && pj.PackageManager != "" {
				return strings.Split(pj.PackageManager, "@")[0]
			}
		}
		return "npm"
	}
}

func detectNodeFramework(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return ""
	}
	var pj struct {
		Dependencies map[string]string `json:"dependencies"`
	}
	if json.Unmarshal(data, &pj) != nil {
		return ""
	}
	for _, fw := range []string{"next", "nuxt", "vite", "react", "vue", "svelte", "express", "fastify", "nestjs", "astro"} {
		for dep := range pj.Dependencies {
			if strings.Contains(dep, fw) {
				switch fw {
				case "next":
					return "Next.js"
				case "nuxt":
					return "Nuxt"
				default:
					return strings.Title(fw)
				}
			}
		}
	}
	return ""
}

func gitBranch(dir string) string {
	head, err := os.ReadFile(filepath.Join(dir, ".git", "HEAD"))
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(string(head))
	if strings.HasPrefix(s, "ref: refs/heads/") {
		return strings.TrimPrefix(s, "ref: refs/heads/")
	}
	return ""
}

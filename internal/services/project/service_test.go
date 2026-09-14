package project

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestInspectGoProject(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/school-api\n\ngo 1.24\n")
	writeFile(t, dir, "Dockerfile", "FROM golang:1.24\n")
	svc := New()
	info, err := svc.Inspect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(info.Languages, "Go") {
		t.Errorf("languages = %v, want Go", info.Languages)
	}
	if !contains(info.Files, "go.mod") {
		t.Errorf("files = %v, want go.mod", info.Files)
	}
}

func TestInspectNodeProject(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{"name":"web","dependencies":{"next":"15.0.0"}}`)
	writeFile(t, dir, "pnpm-lock.yaml", "")
	svc := New()
	info, err := svc.Inspect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(info.Languages, "Node.js") {
		t.Errorf("languages = %v", info.Languages)
	}
	if info.PackageManager != "pnpm" {
		t.Errorf("packageManager = %q, want pnpm", info.PackageManager)
	}
	if info.Framework != "Next.js" {
		t.Errorf("framework = %q, want Next.js", info.Framework)
	}
}

func TestInspectPHPProject(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "composer.json", `{"name":"app/laravel"}`)
	svc := New()
	info, err := svc.Inspect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(info.Languages, "PHP") {
		t.Errorf("languages = %v", info.Languages)
	}
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

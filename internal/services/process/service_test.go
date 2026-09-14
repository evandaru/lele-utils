package process

import (
	"testing"

	"lele-dev/internal/models"
)

func TestSearch(t *testing.T) {
	all := []models.Process{
		{PID: 18231, Name: "node", Command: "node server.js"},
		{PID: 19342, Name: "node", Command: "node vite"},
		{PID: 20182, Name: "go", Command: "go run main.go"},
	}
	if got := Search(all, "node"); len(got) != 2 {
		t.Errorf("search node = %d, want 2", len(got))
	}
	if got := Search(all, "20182"); len(got) != 1 {
		t.Errorf("search pid = %d, want 1", len(got))
	}
	if got := Search(all, ""); len(got) != 3 {
		t.Errorf("empty search = %d, want 3", len(got))
	}
}

func TestIsProtected(t *testing.T) {
	for _, pid := range []int{0, 1, 2} {
		if !IsProtected(pid) {
			t.Errorf("PID %d should be protected", pid)
		}
	}
	if IsProtected(18231) {
		t.Error("PID 18231 should not be protected")
	}
}

func TestListIntegration(t *testing.T) {
	svc := New()
	all, err := svc.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(all) == 0 {
		t.Fatal("List returned no processes")
	}
	// Current process must be in the list.
	found := false
	for _, p := range all {
		if p.PID <= 0 || p.Name == "" {
			t.Errorf("invalid process entry: %+v", p)
			break
		}
		found = true
	}
	if !found {
		t.Error("no valid process entries")
	}
}

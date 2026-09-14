package network

import "testing"

func TestInfoIntegration(t *testing.T) {
	svc := New()
	info, err := svc.Info()
	if err != nil {
		t.Fatalf("Info failed: %v", err)
	}
	if len(info.Interfaces) == 0 {
		t.Fatal("no interfaces detected")
	}
	// Loopback harus selalu ada, tanpa asumsi nama (lo / lo0).
	foundLoopback := false
	for _, ni := range info.Interfaces {
		if ni.Name == "" {
			t.Error("interface with empty name")
		}
		if ni.Type == "loopback" {
			foundLoopback = true
		}
	}
	if !foundLoopback {
		t.Error("no loopback interface detected")
	}
}

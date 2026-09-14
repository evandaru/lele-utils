package shared

import "testing"

func TestSanitizeEnv(t *testing.T) {
	if got := SanitizeEnv("API_KEY", "secret123"); got == "secret123" {
		t.Error("API_KEY should be hidden")
	}
	if got := SanitizeEnv("PATH", "/usr/bin"); got != "/usr/bin" {
		t.Errorf("PATH should pass through, got %q", got)
	}
	for _, k := range []string{"TOKEN", "PASSWORD", "SECRET", "DB_PASSWORD"} {
		if got := SanitizeEnv(k, "x"); got == "x" {
			t.Errorf("%s should be hidden", k)
		}
	}
}

func TestFilter(t *testing.T) {
	items := []string{"18231 node server", "19342 node vite", "20182 go api"}
	if got := Filter(items, "node"); len(got) != 2 {
		t.Errorf("filter node = %d, want 2", len(got))
	}
	if got := Filter(items, ""); len(got) != 3 {
		t.Errorf("empty filter = %d, want 3", len(got))
	}
}

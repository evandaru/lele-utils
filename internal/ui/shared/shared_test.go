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

func TestFormatBytes(t *testing.T) {
	for in, want := range map[int64]string{
		0: "0 B", 512: "512 B", 1024: "1.0 KB",
		524 * 1024: "524.0 KB", 55 * 1024 * 1024: "55.0 MB",
		2 * 1024 * 1024 * 1024: "2.0 GB", -1: "—",
	} {
		if got := FormatBytes(in); got != want {
			t.Errorf("FormatBytes(%d) = %q, want %q", in, got, want)
		}
	}
}

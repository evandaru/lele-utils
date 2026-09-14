package runtime

import "testing"

func TestParseGoVersion(t *testing.T) {
	input := "go version go1.25.1 linux/amd64"
	if got := ParseGoVersion(input); got != "1.25.1" {
		t.Fatalf("ParseGoVersion(%q) = %q, want 1.25.1", input, got)
	}
}

func TestVersionParsers(t *testing.T) {
	cases := []struct {
		name  string
		fn    func(string) string
		input string
		want  string
	}{
		{"node", ParseNodeVersion, "v24.8.0", "24.8.0"},
		{"bun", ParseGenericVersion, "1.2.21", "1.2.21"},
		{"deno", ParseDenoVersion, "deno 2.1.4 (stable, release, x86_64-unknown-linux-gnu)", "2.1.4"},
		{"python", ParsePythonVersion, "Python 3.13.7", "3.13.7"},
		{"php", ParsePHPVersion, "PHP 8.4.15 (cli) (built: ...)", "8.4.15"},
		{"ruby", ParseRubyVersion, "ruby 3.4.5 (2025-07-16 revision ...)", "3.4.5"},
		{"rust", ParseRustVersion, "rustc 1.89.0 (29483883e 2025-08-04)", "1.89.0"},
		{"java-quoted", ParseJavaVersion, "openjdk version \"17.0.16\" 2025-07-15", "17.0.16"},
		{"java-bare", ParseJavaVersion, "openjdk 25 2025-09-16", "25"},
		{"kotlin", ParseKotlinVersion, "Kotlin version 2.1.20-release-394", "2.1.20"},
		{"dart", ParseDartVersion, "Dart SDK version: 3.8.1 (stable)", "3.8.1"},
		{"flutter", ParseFlutterVersion, "Flutter 3.32.5 • channel stable", "3.32.5"},
		{"swift", ParseSwiftVersion, "Swift version 6.1.2 (swift-6.1.2-RELEASE)", "6.1.2"},
		{"gcc", ParseGCCVersion, "gcc (GCC) 14.2.1 20240801", "14.2.1"},
		{"clang", ParseClangVersion, "clang version 18.1.3", "18.1.3"},
	}
	for _, tc := range cases {
		if got := tc.fn(tc.input); got != tc.want {
			t.Errorf("%s: got %q, want %q (input %q)", tc.name, got, tc.want, tc.input)
		}
	}
}

func TestDetectAllReturnsAllRuntimes(t *testing.T) {
	svc := New()
	got := svc.DetectAll()
	if len(got) != len(specs()) {
		t.Fatalf("DetectAll returned %d runtimes, want %d", len(got), len(specs()))
	}
}

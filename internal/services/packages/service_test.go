package packages

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseNpmLsJSON(t *testing.T) {
	data := []byte(`{"dependencies":{"npm":{"version":"10.8.2"},"typescript":{"version":"5.4.5"},"empty":{}}}`)
	pkgs := ParseNpmLsJSON(data)
	if len(pkgs) != 3 {
		t.Fatalf("got %d, want 3", len(pkgs))
	}
	byName := map[string]string{}
	for _, p := range pkgs {
		byName[p.Name] = p.Version
	}
	if byName["typescript"] != "5.4.5" {
		t.Errorf("typescript = %q", byName["typescript"])
	}
	if byName["empty"] != "—" {
		t.Errorf("missing version should be —, got %q", byName["empty"])
	}
	if ParseNpmLsJSON([]byte("not json")) != nil {
		t.Error("invalid JSON should return nil")
	}
}

func TestParsePipListAndShow(t *testing.T) {
	pkgs := ParsePipListJSON([]byte(`[{"name":"requests","version":"2.31.0"},{"name":"pip","version":"24.0"}]`))
	if len(pkgs) != 2 || pkgs[0].Version != "2.31.0" {
		t.Fatalf("got %+v", pkgs)
	}
	show := "Name: requests\nVersion: 2.31.0\nLocation: /usr/lib/python3.12/site-packages\n---\nName: pip\nVersion: 24.0\nLocation: /usr/lib/python3.12/site-packages\n"
	locs := ParsePipShow(show)
	if locs["requests"] != "/usr/lib/python3.12/site-packages" {
		t.Errorf("got %v", locs)
	}
	// pipGuessDir toleran terhadap -/_
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "my_pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := pipGuessDir(dir, "my-pkg"); got == "" {
		t.Error("should guess my_pkg for my-pkg")
	}
}

func TestParseBrewVersions(t *testing.T) {
	pkgs := ParseBrewVersions("node 22.14.0\nwget 1.21.4 1.21.3\nnoversion\n")
	if len(pkgs) != 2 {
		t.Fatalf("got %+v", pkgs)
	}
	if pkgs[0].Version != "22.14.0" {
		t.Errorf("node = %q", pkgs[0].Version)
	}
	if pkgs[1].Version != "1.21.4 (+1)" {
		t.Errorf("wget = %q", pkgs[1].Version)
	}
}

func TestParsePacmanQ(t *testing.T) {
	pkgs := ParsePacmanQ("linux 6.9.1.arch1-1\nnodejs 22.2.0-1\n\nbadline\n", ManagerPacman)
	if len(pkgs) != 2 {
		t.Fatalf("got %+v", pkgs)
	}
	if pkgs[0].Manager != ManagerPacman || pkgs[1].Name != "nodejs" {
		t.Errorf("got %+v", pkgs)
	}
}

func TestParsePacmanQiAndSize(t *testing.T) {
	qi := "Name            : nodejs\nVersion         : 22.2.0-1\nInstalled Size  : 55.10 MiB\n\nName : foo\nVersion : 1.0\nInstalled Size : 225.00 KiB\n"
	sizes := ParsePacmanQi(qi)
	mib := float64(1024 * 1024) // variabel agar konversi runtime (bukan konstanta)
	wantNode := int64(55.10 * mib)
	if sizes["nodejs"] != wantNode {
		t.Errorf("nodejs = %d, want %d", sizes["nodejs"], wantNode)
	}
	if sizes["foo"] != 225*1024 {
		t.Errorf("foo = %d", sizes["foo"])
	}
	for in, want := range map[string]int64{
		"512 B": 512, "1.00 KiB": 1024, "2.50 MiB": 2621440,
		"1.20 GiB": 1288490188, "bogus": 0,
	} {
		if got := ParsePacmanSize(in); got != want {
			t.Errorf("ParsePacmanSize(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestParseDpkgQuery(t *testing.T) {
	pkgs := ParseDpkgQuery("adduser\t3.118\t524\ncurl\t8.6.0-2\t520\nnoversion\t\t100\n")
	if len(pkgs) != 3 {
		t.Fatalf("got %+v", pkgs)
	}
	if pkgs[0].SizeBytes != 524*1024 {
		t.Errorf("adduser size = %d", pkgs[0].SizeBytes)
	}
	if pkgs[2].Version != "—" {
		t.Errorf("empty version = %q", pkgs[2].Version)
	}
}

func TestParseCargoList(t *testing.T) {
	entries := ParseCargoList("cargo-edit v0.13.7:\n    cargo-add\n    cargo-rm\nripgrep v14.1.0:\n    rg\n")
	if len(entries) != 2 {
		t.Fatalf("got %+v", entries)
	}
	if entries[0].name != "cargo-edit" || entries[0].version != "0.13.7" || len(entries[0].bins) != 2 {
		t.Errorf("got %+v", entries[0])
	}
}

func TestParseGemList(t *testing.T) {
	entries := ParseGemList("rake (13.3.0, 12.3.3)\nbundler (default: 2.5.11)\nbadline\n")
	if len(entries) != 2 {
		t.Fatalf("got %+v", entries)
	}
	if entries[0].version != "13.3.0, 12.3.3" || entries[0].firstVer != "13.3.0" {
		t.Errorf("got %+v", entries[0])
	}
	if entries[1].version != "2.5.11" {
		t.Errorf("got %+v", entries[1])
	}
}

func TestParseGoBins(t *testing.T) {
	pkgs := ParseGoBins([]GoBin{{Name: "air", Size: 1024 * 1024}})
	if len(pkgs) != 1 || pkgs[0].Manager != ManagerGo || pkgs[0].Version != "—" {
		t.Fatalf("got %+v", pkgs)
	}
	if pkgs[0].Size != "1.0 MB" {
		t.Errorf("size = %q", pkgs[0].Size)
	}
}

func TestIsSupported(t *testing.T) {
	for _, m := range Order {
		if !IsSupported(m) {
			t.Errorf("%s should be supported", m)
		}
	}
	if IsSupported("apt-get") {
		t.Error("apt-get should not be supported")
	}
}

func TestListUnknownManager(t *testing.T) {
	if _, err := New().List("apt-get"); err == nil {
		t.Error("expected error for unknown manager")
	}
}

func TestListAllIntegration(t *testing.T) {
	pkgs, err := New().List("")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	// Mesin boleh minim, tapi hasil harus valid bila ada.
	for _, p := range pkgs {
		if p.Name == "" || !IsSupported(p.Manager) || p.Size == "" {
			t.Errorf("invalid package: %+v", p)
			break
		}
	}
	t.Logf("%d packages detected", len(pkgs))
}

func TestSearch(t *testing.T) {
	pkgs := ParsePacmanQ("nodejs 22.2.0-1\npython 3.12.3-1\n", ManagerPacman)
	if got := Search(pkgs, "node"); len(got) != 1 {
		t.Errorf("search node = %d, want 1", len(got))
	}
	if got := Search(pkgs, "pacman"); len(got) != 2 {
		t.Errorf("search manager = %d, want 2", len(got))
	}
}

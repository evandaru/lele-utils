package container

import "testing"

func TestParsePsOutput(t *testing.T) {
	out := "abc123\tpostgres\tpostgres:17\tUp 2 hours\t\t0.0.0.0:5432->5432/tcp\n" +
		"def456\tredis\tredis:8\tUp 2 hours\t\t0.0.0.0:6379->6379/tcp\n"
	cs := ParsePsOutput("docker", out)
	if len(cs) != 2 {
		t.Fatalf("got %d containers, want 2", len(cs))
	}
	if cs[0].Name != "postgres" || cs[0].Image != "postgres:17" || cs[0].Runtime != "docker" {
		t.Errorf("first = %+v", cs[0])
	}
}

func TestParsePsOutputEmpty(t *testing.T) {
	if cs := ParsePsOutput("docker", ""); len(cs) != 0 {
		t.Errorf("got %d, want 0", len(cs))
	}
}

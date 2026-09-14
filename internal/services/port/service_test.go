package port

import "testing"

const sampleSs = `State  Recv-Q Send-Q Local Address:Port Peer Address:Port Process
LISTEN 0      128    127.0.0.1:3000      0.0.0.0:*     users:(("node",pid=18231,fd=21))
LISTEN 0      128    127.0.0.1:5173      0.0.0.0:*     users:(("node",pid=19342,fd=22))
LISTEN 0      128    127.0.0.1:5432      0.0.0.0:*     users:(("postgres",pid=912,fd=7))
`

const sampleLsof = `COMMAND   PID  USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
node    18231 pojan  21u  IPv4 123456      0t0  TCP 127.0.0.1:3000 (LISTEN)
postgres  912 postgres 7u IPv4 123457      0t0  TCP 127.0.0.1:5432 (LISTEN)
`

func TestParseSsOutput(t *testing.T) {
	ports := ParseSsOutput(sampleSs)
	if len(ports) != 3 {
		t.Fatalf("got %d ports, want 3", len(ports))
	}
	if ports[0].Port != 3000 || ports[0].PID != 18231 || ports[0].ProcessName != "node" {
		t.Errorf("first port = %+v, want 3000/node/18231", ports[0])
	}
	if ports[2].Port != 5432 || ports[2].ProcessName != "postgres" {
		t.Errorf("third port = %+v, want 5432/postgres", ports[2])
	}
}

func TestParseLsofOutput(t *testing.T) {
	ports := ParseLsofOutput(sampleLsof)
	if len(ports) != 2 {
		t.Fatalf("got %d ports, want 2", len(ports))
	}
	if ports[0].Port != 3000 || ports[0].PID != 18231 {
		t.Errorf("first port = %+v", ports[0])
	}
}

func TestSearch(t *testing.T) {
	all := ParseSsOutput(sampleSs)
	if got := Search(all, "3000"); len(got) != 1 {
		t.Errorf("search 3000 = %d results, want 1", len(got))
	}
	if got := Search(all, "node"); len(got) != 2 {
		t.Errorf("search node = %d results, want 2", len(got))
	}
	if got := Search(all, ""); len(got) != 3 {
		t.Errorf("empty search = %d results, want 3", len(got))
	}
}

func TestLocalhostURL(t *testing.T) {
	if got := LocalhostURL(3000); got != "http://localhost:3000" {
		t.Errorf("got %q", got)
	}
}

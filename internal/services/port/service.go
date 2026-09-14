package port

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"dev-utils/internal/models"

	gnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// Service — PRD §11. Port → PID → process.
type Service struct{}

func New() *Service { return &Service{} }

// List returns listening TCP ports. Strategi:
// 1. gopsutil connections (lintas platform, tanpa asumsi tool).
// 2. Enrich process name via gopsutil process.
// 3. Jika PID 0 (kurang permission), coba perkaya via ss/lsof.
func (s *Service) List() ([]models.Port, error) {
	conns, err := gnet.Connections("tcp")
	if err != nil {
		return listViaOSCommands()
	}
	seen := map[string]*models.Port{}
	for _, c := range conns {
		if !strings.EqualFold(c.Status, "LISTEN") {
			continue
		}
		proto := "TCP"
		if c.Type == 2 { // SOCK_DGRAM
			proto = "UDP"
		}
		addr := ""
		if c.Laddr.IP != "" {
			addr = c.Laddr.IP
		}
		key := fmt.Sprintf("%s/%d/%s", proto, c.Laddr.Port, addr)
		if _, ok := seen[key]; ok {
			continue
		}
		p := &models.Port{
			Port:     int(c.Laddr.Port),
			Protocol: proto,
			Address:  addr,
			PID:      int(c.Pid),
		}
		seen[key] = p
	}
	out := make([]models.Port, 0, len(seen))
	for _, p := range seen {
		enrich(p)
		out = append(out, *p)
	}
	// Fallback: jika kosong (mis. permission), coba ss/lsof.
	if len(out) == 0 {
		if fb, err := listViaOSCommands(); err == nil && len(fb) > 0 {
			return fb, nil
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Port < out[j].Port })
	return out, nil
}

func enrich(p *models.Port) {
	if p.PID > 0 {
		if proc, err := process.NewProcess(int32(p.PID)); err == nil {
			if n, err := proc.Name(); err == nil {
				p.ProcessName = n
			}
			if cmd, err := proc.Cmdline(); err == nil {
				p.Command = cmd
			}
		}
	}
	if p.ProcessName == "" {
		p.ProcessName = lookupViaOS(p.Port, p.Protocol)
	}
}

// Find returns detail for a single port.
func (s *Service) Find(port int) (*models.Port, error) {
	all, err := s.List()
	if err != nil {
		return nil, err
	}
	for _, p := range all {
		if p.Port == port {
			cp := p
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("port %d tidak sedang listening", port)
}

// Search filters by port number / process / address (PRD §18).
func Search(all []models.Port, query string) []models.Port {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return all
	}
	var out []models.Port
	for _, p := range all {
		if strings.Contains(fmt.Sprint(p.Port), q) ||
			strings.Contains(strings.ToLower(p.ProcessName), q) ||
			strings.Contains(strings.ToLower(p.Address), q) {
			out = append(out, p)
		}
	}
	return out
}

// LocalhostURL builds http://localhost:port (PRD §11 shortcut `o`).
func LocalhostURL(port int) string { return fmt.Sprintf("http://localhost:%d", port) }

// ---- OS command fallback (ss → lsof → netstat) ----

func listViaOSCommands() ([]models.Port, error) {
	if out, err := exec.Command("ss", "-tlnp").Output(); err == nil {
		if ports := ParseSsOutput(string(out)); len(ports) > 0 {
			return ports, nil
		}
	}
	if out, err := exec.Command("lsof", "-i", "-P", "-n", "-sTCP:LISTEN").Output(); err == nil {
		if ports := ParseLsofOutput(string(out)); len(ports) > 0 {
			return ports, nil
		}
	}
	return nil, fmt.Errorf("tidak dapat mendeteksi listening ports (butuh permission atau tool ss/lsof)")
}

func lookupViaOS(port int, proto string) string {
	if _, err := exec.LookPath("ss"); err == nil {
		if out, err := exec.Command("ss", "-tlnp").Output(); err == nil {
			for _, p := range ParseSsOutput(string(out)) {
				if p.Port == port {
					return p.ProcessName
				}
			}
		}
	}
	return ""
}

// ParseSsOutput parses `ss -tlnp` output. Pure function → unit tested.
func ParseSsOutput(out string) []models.Port {
	var ports []models.Port
	lines := strings.Split(out, "\n")
	rePort := regexp.MustCompile(`(?:\[?([0-9a-fA-F:.]+)\]?:|\*:)(\d+)\s*$`)
	_ = rePort
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "State") || strings.HasPrefix(line, "Netid") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		// Kolom berbeda tergantung flag ss: `ss -tlnp` → [tcp LISTEN ... LOCAL PEER ...],
		// tanpa Netid → [LISTEN ... LOCAL PEER ...]. Cari field addr:port numerik pertama.
		portNum, addr := 0, ""
		for _, f := range fields {
			if n, a := splitAddrPort(f); n != 0 {
				// hindari kolom Peer 0.0.0.0:* (port 0) — splitAddrPort sudah return 0 untuk itu
				portNum, addr = n, a
				break
			}
		}
		if portNum == 0 {
			continue
		}
		pid := 0
		procName := ""
		// users:(("node",pid=18231,fd=3))
		if m := regexp.MustCompile(`"([^"]+)",pid=(\d+)`).FindStringSubmatch(line); m != nil {
			procName = m[1]
			pid, _ = strconv.Atoi(m[2])
		}
		ports = append(ports, models.Port{
			Port:        portNum,
			Protocol:    "TCP",
			Address:     addr,
			PID:         pid,
			ProcessName: procName,
		})
	}
	return ports
}

// ParseLsofOutput parses `lsof -i -P -n -sTCP:LISTEN`. Pure → unit tested.
func ParseLsofOutput(out string) []models.Port {
	var ports []models.Port
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // header
		}
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}
		procName := fields[0]
		pid, _ := strconv.Atoi(fields[1])
		addrPort := fields[8] // e.g. 127.0.0.1:3000 atau *:8080
		portNum, addr := splitAddrPort(addrPort)
		if portNum == 0 {
			continue
		}
		proto := "TCP"
		if len(fields) > 7 && strings.Contains(strings.ToUpper(fields[7]), "UDP") {
			proto = "UDP"
		}
		ports = append(ports, models.Port{
			Port:        portNum,
			Protocol:    proto,
			Address:     addr,
			PID:         pid,
			ProcessName: procName,
		})
	}
	return ports
}

func splitAddrPort(s string) (int, string) {
	// Tangani IPv6 [::]:8080, *:3000, 127.0.0.1:3000
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "[") {
		if h, p, err := net.SplitHostPort(s); err == nil {
			n, _ := strconv.Atoi(p)
			return n, h
		}
	}
	// *:port
	if strings.HasPrefix(s, "*:") {
		n, _ := strconv.Atoi(strings.TrimPrefix(s, "*:"))
		return n, "*"
	}
	if h, p, err := net.SplitHostPort(s); err == nil {
		n, _ := strconv.Atoi(p)
		return n, h
	}
	// fallback: ambil setelah ':' terakhir
	if idx := strings.LastIndex(s, ":"); idx >= 0 {
		n, err := strconv.Atoi(s[idx+1:])
		if err == nil {
			return n, s[:idx]
		}
	}
	return 0, ""
}

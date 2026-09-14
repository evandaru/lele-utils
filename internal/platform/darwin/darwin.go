package darwin

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"

	"dev-utils/internal/models"

	"github.com/shirou/gopsutil/v3/process"
)

// ListProcesses via gopsutil (sama reliable di macOS).
func ListProcesses() ([]models.Process, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("gagal membaca daftar process: %w", err)
	}
	out := make([]models.Process, 0, len(procs))
	for _, p := range procs {
		name, _ := p.Name()
		cmd, _ := p.Cmdline()
		username, _ := p.Username()
		cpu, _ := p.CPUPercent()
		mem, _ := p.MemoryInfo()
		var memMB float64
		if mem != nil {
			memMB = float64(mem.RSS) / 1024 / 1024
		}
		if name == "" && cmd == "" {
			continue
		}
		if name == "" {
			name = shortCmd(cmd)
		}
		out = append(out, models.Process{
			PID:        int(p.Pid),
			Name:       name,
			Command:    cmd,
			User:       username,
			CPUPercent: cpu,
			MemoryMB:   memMB,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PID < out[j].PID })
	return out, nil
}

func shortCmd(cmd string) string {
	if cmd == "" {
		return "?"
	}
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "?"
	}
	base := parts[0]
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[i+1:]
	}
	return base
}

// NetworkInfo — interfaces dari stdlib (tanpa asumsi en0/lo0),
// gateway via `route -n get default`, DNS via /etc/resolv.conf + scutil fallback.
func NetworkInfo() (*models.NetworkInfo, error) {
	info := &models.NetworkInfo{}
	if h, err := os.Hostname(); err == nil {
		info.Hostname = h
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("gagal membaca network interfaces: %w", err)
	}
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		ip := ""
		for _, a := range addrs {
			s := strings.Split(a.String(), "/")[0]
			parsed := net.ParseIP(s)
			if parsed != nil && parsed.To4() != nil {
				ip = s
				break
			}
		}
		if ip == "" && len(addrs) > 0 {
			ip = strings.Split(addrs[0].String(), "/")[0]
		}
		status := "DOWN"
		if iface.Flags&net.FlagUp != 0 {
			status = "UP"
		}
		typ := "ethernet"
		switch {
		case strings.HasPrefix(iface.Name, "lo"):
			typ = "loopback"
		case strings.HasPrefix(iface.Name, "en"):
			typ = "wifi/ethernet"
		case strings.HasPrefix(iface.Name, "bridge"):
			typ = "bridge"
		}
		info.Interfaces = append(info.Interfaces, models.NetworkInterface{
			Name:    iface.Name,
			Address: ip,
			Type:    typ,
			Status:  status,
			MAC:     iface.HardwareAddr.String(),
		})
	}
	info.Gateway = defaultGateway()
	info.DNS = resolvDNS()
	return info, nil
}

func defaultGateway() string {
	out, err := exec.Command("route", "-n", "get", "default").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "gateway:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "gateway:"))
		}
	}
	return ""
}

func resolvDNS() []string {
	f, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "nameserver") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				out = append(out, parts[1])
			}
		}
	}
	return out
}

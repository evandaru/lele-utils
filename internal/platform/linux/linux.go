package linux

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/user"
	"sort"
	"strings"

	"lele-dev/internal/models"

	"github.com/shirou/gopsutil/v3/process"
)

// List processes via gopsutil (lintas platform, reliable).
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
		if username == "" {
			if u, err := user.Current(); err == nil {
				_ = u
			}
		}
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

// NetworkInfo — interfaces dari stdlib, gateway dari /proc/net/route, DNS dari /etc/resolv.conf.
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
			var s string
			switch v := a.(type) {
			case *net.IPNet:
				if v.IP.To4() != nil {
					s = v.IP.String()
				}
			case *net.IPAddr:
				if v.IP.To4() != nil {
					s = v.IP.String()
				}
			}
			if s != "" {
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
		case strings.HasPrefix(iface.Name, "wl"):
			typ = "wifi"
		case strings.HasPrefix(iface.Name, "docker"):
			typ = "bridge"
		case strings.HasPrefix(iface.Name, "br-") || strings.HasPrefix(iface.Name, "bridge"):
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
	f, err := os.Open("/proc/net/route")
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 3 || fields[0] == "Iface" {
			continue
		}
		// destination 00000000 + mask 00000000 = default route
		if fields[1] == "00000000" {
			return hexLEToIP(fields[2])
		}
	}
	return ""
}

func hexLEToIP(hex string) string {
	if len(hex) != 8 {
		return ""
	}
	var b [4]byte
	for i := 0; i < 4; i++ {
		var v unsigned
		_, _ = fmt.Sscanf(hex[i*2:i*2+2], "%02x", &v)
		b[i] = byte(v)
	}
	// /proc/net/route menyimpan gateway little-endian
	return fmt.Sprintf("%d.%d.%d.%d", b[3], b[2], b[1], b[0])
}

type unsigned uint

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

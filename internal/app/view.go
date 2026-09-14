package app

import (
	"fmt"
	"os"
	"strings"

	"dev-utils/internal/ui/shared"
)

// View renders current screen. Gaya: minimal, dense, keyboard-first (PRD §26).
func (m Model) View() string {
	if m.errMsg != "" {
		return m.renderError()
	}
	if m.confirm != nil {
		return m.renderConfirm()
	}
	if m.logsView != "" {
		return m.renderLogs()
	}
	var body string
	switch m.screen {
	case ScreenProcesses:
		body = m.viewProcesses()
	case ScreenProcessDetail:
		body = m.viewProcessDetail()
	case ScreenPorts:
		body = m.viewPorts()
	case ScreenPortDetail:
		body = m.viewPortDetail()
	case ScreenRuntimes:
		body = m.viewRuntimes()
	case ScreenNetwork:
		body = m.viewNetwork()
	case ScreenTools:
		body = m.viewTools()
	case ScreenContainers:
		body = m.viewContainers()
	case ScreenProjects:
		body = m.viewProjects()
	case ScreenSystem:
		body = m.viewSystem()
	default:
		body = m.viewDashboard()
	}
	if m.showHelp {
		body += "\n" + m.viewHelp()
	}
	if m.searching {
		body += "\n" + m.search.View()
	} else if m.query != "" {
		body += "\n" + shared.Dim.Render(fmt.Sprintf("/ %s   (esc clears filter)", m.query))
	}
	if m.message != "" {
		body += "\n" + shared.Green.Render(m.message)
	}
	if w := rootWarning(); w != "" && m.screen == ScreenDashboard {
		body += "\n" + w
	}
	return body
}

func (m Model) header(title string) string {
	return shared.Header.Render(" "+title+" ") + "  " + shared.Dim.Render("dev-utils v0.1.0   esc=back  q=quit")
}

func (m Model) renderError() string {
	return shared.Box.Render(
		shared.Red.Render("Unable to complete request.") + "\n\n" +
			m.errMsg + "\n",
	)
}

func (m Model) renderConfirm() string {
	c := m.confirm
	var b strings.Builder
	if c.kind == "port-process" {
		b.WriteString(shared.Title.Render(fmt.Sprintf("Kill process pada %s?", c.label)) + "\n\n")
	} else {
		b.WriteString(shared.Title.Render(fmt.Sprintf("Kill process %s?", c.label)) + "\n\n")
	}
	if c.stage == 1 {
		b.WriteString("Default action: SIGTERM (graceful).\n\n")
		b.WriteString("[y] Yes   [n] No   [K] Force kill (SIGKILL, konfirmasi kedua)\n")
	} else {
		b.WriteString(shared.Red.Render("⚠ FORCE KILL (SIGKILL) — tidak bisa dibatalkan.") + "\n\n")
		b.WriteString("[y] Yes, force kill   [n] Cancel\n")
	}
	return shared.Box.Render(b.String())
}

func (m Model) renderLogs() string {
	return shared.Box.Render(shared.Trunc(m.logsView, 4000) + "\n\n" + shared.Dim.Render("press any key to close"))
}

// ---- Dashboard (PRD §9) ----

func (m Model) viewDashboard() string {
	var b strings.Builder
	b.WriteString(m.header("dev-utils — Developer Workstation Toolkit") + "\n\n")
	procs := fmt.Sprintf("%d running", len(m.processes))
	ports := fmt.Sprintf("%d listening", len(m.ports))
	nrt := 0
	for _, r := range m.runtimes {
		if r.Installed {
			nrt++
		}
	}
	nif := 0
	if m.netinfo != nil {
		nif = len(m.netinfo.Interfaces)
	}
	ntools := 0
	for _, t := range m.tools {
		if t.Installed {
			ntools++
		}
	}
	if m.loading {
		b.WriteString(shared.Dim.Render("  scanning system (concurrent)...") + "\n\n")
	}
	b.WriteString(fmt.Sprintf("  Processes       %s\n", procs))
	b.WriteString(fmt.Sprintf("  Ports           %s\n", ports))
	b.WriteString(fmt.Sprintf("  Runtimes        %d detected\n", nrt))
	b.WriteString(fmt.Sprintf("  Network         %d interfaces\n", nif))
	b.WriteString(fmt.Sprintf("  Dev Tools       %d detected\n", ntools))
	if m.project != nil && len(m.project.Languages) > 0 {
		b.WriteString(fmt.Sprintf("\n  PROJECT  %s  [%s]\n", m.project.Name, strings.Join(m.project.Languages, ", ")))
	}
	b.WriteString("\n")
	items := []string{"Processes", "Ports", "Runtimes", "Network", "Dev Tools", "Containers", "Projects", "System"}
	descs := []string{"Running processes", "Listening ports", "Programming languages", "Network information", "Installed developer tools", "Docker / Podman", "Detect current project", "CPU / RAM / Disk"}
	for i, it := range items {
		marker := fmt.Sprintf("  %d  %-12s %s", i+1, it, descs[i])
		if i == m.cursor {
			b.WriteString("  " + shared.Selected.Render(marker) + "\n")
		} else {
			b.WriteString(marker + "\n")
		}
	}
	b.WriteString("\n  q  Quit\n")
	b.WriteString("\n" + shared.HelpLine("[1-8] open"))
	return shared.Box.Render(b.String())
}

// ---- Processes (PRD §10) ----

func (m Model) viewProcesses() string {
	var b strings.Builder
	b.WriteString(m.header(fmt.Sprintf("PROCESS MANAGER — %d processes", len(m.processes))) + "\n\n")
	b.WriteString(fmt.Sprintf("  %-8s %-18s %-8s %-10s %s\n", "PID", "NAME", "CPU", "MEMORY", "COMMAND"))
	b.WriteString("  " + strings.Repeat("-", 70) + "\n")
	items := m.filteredProcesses()
	if len(items) == 0 {
		b.WriteString(shared.Dim.Render("  (no processes match)\n"))
	}
	max := 20
	start := 0
	if m.cursor > max-1 {
		start = m.cursor - max + 1
	}
	for i := start; i < len(items) && i < start+max; i++ {
		p := m.processes[items[i]]
		line := fmt.Sprintf("  %-8d %-18s %-8s %-10s %s",
			p.PID, shared.Trunc(p.Name, 18),
			fmt.Sprintf("%.1f%%", p.CPUPercent),
			formatMB(p.MemoryMB),
			shared.Trunc(p.Command, 40))
		if i == m.cursor {
			b.WriteString(shared.Selected.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}
	b.WriteString("\n" + shared.HelpLine("[enter] details  [k] kill  [/] search  [r] refresh"))
	return shared.Box.Render(b.String())
}

func (m Model) viewProcessDetail() string {
	if m.detailIdx < 0 || m.detailIdx >= len(m.processes) {
		return "Process not found."
	}
	p := m.processes[m.detailIdx]
	var b strings.Builder
	b.WriteString(m.header(fmt.Sprintf("PROCESS %d", p.PID)) + "\n\n")
	b.WriteString(fmt.Sprintf("  PID:       %d\n", p.PID))
	b.WriteString(fmt.Sprintf("  Name:      %s\n", p.Name))
	b.WriteString(fmt.Sprintf("  Command:   %s\n", p.Command))
	b.WriteString(fmt.Sprintf("  User:      %s\n", p.User))
	b.WriteString(fmt.Sprintf("  CPU:       %.1f%%\n", p.CPUPercent))
	b.WriteString(fmt.Sprintf("  Memory:    %s\n", formatMB(p.MemoryMB)))
	b.WriteString("\n  Ports:\n")
	found := false
	for _, port := range m.ports {
		if port.PID == p.PID {
			b.WriteString(fmt.Sprintf("    :%d (%s)\n", port.Port, port.Protocol))
			found = true
		}
	}
	if !found {
		b.WriteString(shared.Dim.Render("    (none)\n"))
	}
	b.WriteString("\n" + shared.HelpLine("[k] kill  [esc] back"))
	return shared.Box.Render(b.String())
}

// ---- Ports (PRD §11) ----

func (m Model) viewPorts() string {
	var b strings.Builder
	b.WriteString(m.header(fmt.Sprintf("PORTS — %d listening", len(m.ports))) + "\n\n")
	b.WriteString(fmt.Sprintf("  %-8s %-10s %-8s %-16s %s\n", "PORT", "PROTOCOL", "PID", "ADDRESS", "PROCESS"))
	b.WriteString("  " + strings.Repeat("-", 70) + "\n")
	items := m.filteredPorts()
	if len(items) == 0 {
		b.WriteString(shared.Dim.Render("  (no ports match — coba jalankan dev server / database)\n"))
	}
	max := 20
	start := 0
	if m.cursor > max-1 {
		start = m.cursor - max + 1
	}
	for i := start; i < len(items) && i < start+max; i++ {
		p := m.ports[items[i]]
		line := fmt.Sprintf("  %-8d %-10s %-8d %-16s %s", p.Port, p.Protocol, p.PID, shared.Trunc(p.Address, 16), p.ProcessName)
		if i == m.cursor {
			b.WriteString(shared.Selected.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}
	b.WriteString("\n" + shared.HelpLine("[enter] details  [k] kill  [o] open localhost  [c] copy  [/] search"))
	return shared.Box.Render(b.String())
}

func (m Model) viewPortDetail() string {
	if m.detailIdx < 0 || m.detailIdx >= len(m.ports) {
		return "Port not found."
	}
	p := m.ports[m.detailIdx]
	var b strings.Builder
	b.WriteString(m.header(fmt.Sprintf("PORT %d", p.Port)) + "\n\n")
	b.WriteString(fmt.Sprintf("  Address:    %s\n", p.Address))
	b.WriteString(fmt.Sprintf("  Protocol:   %s\n", p.Protocol))
	b.WriteString(fmt.Sprintf("  PID:        %d\n", p.PID))
	b.WriteString(fmt.Sprintf("  Process:    %s\n", p.ProcessName))
	b.WriteString(fmt.Sprintf("\n  Command:\n  %s\n", p.Command))
	b.WriteString(fmt.Sprintf("\n  URL:        http://localhost:%d\n", p.Port))
	b.WriteString("\n" + shared.HelpLine("[k] kill process  [o] open localhost  [esc] back"))
	return shared.Box.Render(b.String())
}

// ---- Runtimes (PRD §12) ----

func (m Model) viewRuntimes() string {
	var b strings.Builder
	b.WriteString(m.header("RUNTIMES") + "\n\n")
	b.WriteString(fmt.Sprintf("  %-12s %-14s %s\n", "NAME", "VERSION", "PATH"))
	b.WriteString("  " + strings.Repeat("-", 60) + "\n")
	for _, r := range m.runtimes {
		ver := r.Version
		path := r.Path
		if !r.Installed {
			ver = shared.Dim.Render("—")
			path = shared.Dim.Render("not installed")
		}
		b.WriteString(fmt.Sprintf("  %-12s %-14s %s\n", r.Name, ver, path))
	}
	b.WriteString("\n" + shared.HelpLine("[r] refresh"))
	return shared.Box.Render(b.String())
}

// ---- Network (PRD §13) ----

func (m Model) viewNetwork() string {
	var b strings.Builder
	b.WriteString(m.header("NETWORK") + "\n\n")
	if m.netinfo == nil {
		b.WriteString(shared.Dim.Render("  (loading...)"))
		return shared.Box.Render(b.String())
	}
	b.WriteString("  Interfaces\n\n")
	b.WriteString(fmt.Sprintf("  %-12s %-16s %-8s %s\n", "NAME", "ADDRESS", "STATUS", "TYPE"))
	b.WriteString("  " + strings.Repeat("-", 50) + "\n")
	for _, ni := range m.netinfo.Interfaces {
		b.WriteString(fmt.Sprintf("  %-12s %-16s %-8s %s\n", ni.Name, ni.Address, ni.Status, ni.Type))
	}
	b.WriteString(fmt.Sprintf("\n  Gateway\n  %s\n", nonEmpty(m.netinfo.Gateway, "(unknown)")))
	b.WriteString("\n  DNS\n")
	if len(m.netinfo.DNS) == 0 {
		b.WriteString("  (unknown)\n")
	}
	for _, d := range m.netinfo.DNS {
		b.WriteString(fmt.Sprintf("  %s\n", d))
	}
	if m.netinfo.Hostname != "" {
		b.WriteString(fmt.Sprintf("\n  Hostname\n  %s\n", m.netinfo.Hostname))
	}
	b.WriteString("\n" + shared.HelpLine("[r] refresh"))
	return shared.Box.Render(b.String())
}

// ---- Tools (PRD §14) ----

func (m Model) viewTools() string {
	var b strings.Builder
	b.WriteString(m.header("DEV TOOLS") + "\n\n")
	b.WriteString(fmt.Sprintf("  %-12s %-14s %-6s %s\n", "TOOL", "VERSION", "STATUS", "PATH"))
	b.WriteString("  " + strings.Repeat("-", 60) + "\n")
	for _, t := range m.tools {
		status := shared.Red.Render("✗")
		ver := shared.Dim.Render("—")
		path := shared.Dim.Render("not installed")
		if t.Installed {
			status = shared.Green.Render("✓")
			ver = t.Version
			path = t.Path
		}
		b.WriteString(fmt.Sprintf("  %-12s %-14s %-6s %s\n", t.Name, ver, status, shared.Trunc(path, 30)))
	}
	b.WriteString("\n" + shared.HelpLine("[r] refresh"))
	return shared.Box.Render(b.String())
}

// ---- Containers (PRD §15) ----

func (m Model) viewContainers() string {
	var b strings.Builder
	b.WriteString(m.header(fmt.Sprintf("CONTAINERS — %d running", len(m.containers))) + "\n\n")
	if len(m.containers) == 0 {
		b.WriteString("  Docker: " + dockerStatus("docker") + "\n")
		b.WriteString("  Podman: " + dockerStatus("podman") + "\n\n")
		b.WriteString(shared.Dim.Render("  (no containers — start docker/podman atau jalankan container)\n"))
		b.WriteString("\n" + shared.HelpLine("[r] refresh"))
		return shared.Box.Render(b.String())
	}
	b.WriteString(fmt.Sprintf("  %-16s %-20s %-10s %-10s %s\n", "NAME", "IMAGE", "RUNTIME", "STATE", "PORTS"))
	b.WriteString("  " + strings.Repeat("-", 80) + "\n")
	items := m.filteredContainers()
	for i, idx := range items {
		c := m.containers[idx]
		line := fmt.Sprintf("  %-16s %-20s %-10s %-10s %s",
			shared.Trunc(c.Name, 16), shared.Trunc(c.Image, 20), c.Runtime, shared.Trunc(c.State, 10), shared.Trunc(c.Ports, 24))
		if i == m.cursor {
			b.WriteString(shared.Selected.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}
	b.WriteString("\n" + shared.HelpLine("[s] start  [t] stop  [r] restart  [l] logs  [k] kill"))
	return shared.Box.Render(b.String())
}

// ---- Projects (PRD §16) ----

func (m Model) viewProjects() string {
	var b strings.Builder
	b.WriteString(m.header("PROJECT") + "\n\n")
	if m.project == nil {
		b.WriteString(shared.Dim.Render("  (loading...)"))
		return shared.Box.Render(b.String())
	}
	p := m.project
	b.WriteString(fmt.Sprintf("  Name:\n  %s\n\n  Path:\n  %s\n", p.Name, p.Path))
	if len(p.Languages) > 0 {
		b.WriteString("\n  Detected:\n\n")
		for _, l := range p.Languages {
			b.WriteString(fmt.Sprintf("  %s %s\n", shared.Green.Render("✓"), l))
		}
		for _, t := range p.Tools {
			if t == "env" {
				continue
			}
			already := false
			for _, l := range p.Languages {
				if strings.EqualFold(l, t) {
					already = true
				}
			}
			if !already {
				b.WriteString(fmt.Sprintf("  %s %s\n", shared.Green.Render("✓"), t))
			}
		}
	} else {
		b.WriteString(shared.Dim.Render("\n  (no known project markers — bukan Go/Node/PHP/Rust project)\n"))
	}
	if p.PackageManager != "" {
		b.WriteString(fmt.Sprintf("\n  Package manager: %s\n", p.PackageManager))
	}
	if p.Framework != "" {
		b.WriteString(fmt.Sprintf("  Framework: %s\n", p.Framework))
	}
	if p.GitBranch != "" {
		b.WriteString(fmt.Sprintf("  Git branch: %s\n", p.GitBranch))
	}
	if len(p.Files) > 0 {
		b.WriteString("\n  Files:\n\n")
		for _, f := range p.Files {
			b.WriteString(fmt.Sprintf("  %s\n", f))
		}
	}
	b.WriteString("\n" + shared.HelpLine("[r] refresh"))
	return shared.Box.Render(b.String())
}

// ---- System (PRD §17 + env redaction §23) ----

func (m Model) viewSystem() string {
	var b strings.Builder
	b.WriteString(m.header("SYSTEM") + "\n\n")
	if m.sysinfo == nil {
		b.WriteString(shared.Dim.Render("  (loading...)"))
		return shared.Box.Render(b.String())
	}
	s := m.sysinfo
	b.WriteString(fmt.Sprintf("  OS\n  %s (%s)\n\n", s.OS, s.Arch))
	b.WriteString(fmt.Sprintf("  Kernel\n  %s\n\n", s.Kernel))
	b.WriteString(fmt.Sprintf("  CPU\n  %s (%d cores)\n\n", nonEmpty(s.CPUModel, "unknown"), s.CPUCores))
	b.WriteString(fmt.Sprintf("  Memory\n  %.1f GB total, %.1f GB used\n\n", s.MemTotalGB, s.MemUsedGB))
	b.WriteString(fmt.Sprintf("  Disk\n  %.1f GB total, %.1f GB used\n\n", s.DiskTotalGB, s.DiskUsedGB))
	b.WriteString(fmt.Sprintf("  Uptime\n  %s\n", s.Uptime))
	b.WriteString("\n  Environment (secrets hidden)\n\n")
	for _, kv := range []struct{ k, v string }{
		{"PATH", os.Getenv("PATH")},
		{"SHELL", os.Getenv("SHELL")},
		{"EDITOR", os.Getenv("EDITOR")},
	} {
		if kv.v == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("  %-8s %s\n", kv.k, shared.Trunc(shared.SanitizeEnv(kv.k, kv.v), 60)))
	}
	b.WriteString("\n" + shared.HelpLine("[r] refresh"))
	return shared.Box.Render(b.String())
}

func (m Model) viewHelp() string {
	return shared.Box.Render(
		"  GLOBAL  ↑↓ navigate  enter select  esc back  q quit  r refresh  / search  ? help\n" +
			"  PROCESS  k kill (SIGTERM, konfirmasi; K = force SIGKILL)\n" +
			"  PORT     k kill  o open localhost  c copy port\n" +
			"  CONTAINER  s start  t stop  r restart  l logs  k kill",
	)
}

func formatMB(mb float64) string {
	if mb >= 1024 {
		return fmt.Sprintf("%.1f GB", mb/1024)
	}
	return fmt.Sprintf("%.0f MB", mb)
}

func nonEmpty(s, fb string) string {
	if s == "" {
		return fb
	}
	return s
}

func dockerStatus(bin string) string {
	for _, e := range os.Environ() {
		_ = e
	}
	// ringan: cek PATH
	for _, dir := range strings.Split(os.Getenv("PATH"), ":") {
		if _, err := os.Stat(dir + "/" + bin); err == nil {
			return shared.Green.Render("Installed")
		}
	}
	return shared.Dim.Render("Not installed")
}

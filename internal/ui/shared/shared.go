package shared

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	Title    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	Header   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4")).Padding(0, 1)
	Selected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("7"))
	Dim      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	Green    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	Yellow   = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	Red      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	Box      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
)

// HelpLine renders global key hints (PRD §27).
func HelpLine(extra string) string {
	base := "[↑↓] navigate  [enter] select  [esc] back  [q] quit  [r] refresh  [/] search  [?] help"
	if extra != "" {
		base += "  " + extra
	}
	return Dim.Render(base)
}

// Filter returns indices matching query (case-insensitive substring over joined fields).
func Filter(items []string, query string) []int {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		idx := make([]int, len(items))
		for i := range items {
			idx[i] = i
		}
		return idx
	}
	var out []int
	for i, s := range items {
		if strings.Contains(strings.ToLower(s), q) {
			out = append(out, i)
		}
	}
	return out
}

// SanitizeEnv hides secret values (PRD §23.7).
func SanitizeEnv(key, value string) string {
	up := strings.ToUpper(key)
	for _, marker := range []string{"KEY", "TOKEN", "PASSWORD", "SECRET", "CREDENTIAL", "PRIVATE"} {
		if strings.Contains(up, marker) {
			return "•••••••• (hidden)"
		}
	}
	return value
}

// Trunc shortens long strings for dense tables.
func Trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}

// FormatBytes renders byte counts human-readable (B/KB/MB/GB).
func FormatBytes(b int64) string {
	if b < 0 {
		return "—"
	}
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	f := float64(b) / 1024
	for _, u := range []string{"KB", "MB", "GB"} {
		if f < 1024 || u == "GB" {
			return fmt.Sprintf("%.1f %s", f, u)
		}
		f /= 1024
	}
	return fmt.Sprintf("%.1f GB", f)
}

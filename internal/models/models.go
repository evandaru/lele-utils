package models

// Process — PRD §8.
type Process struct {
	PID        int     `json:"pid"`
	Name       string  `json:"name"`
	Command    string  `json:"command"`
	User       string  `json:"user"`
	CPUPercent float64 `json:"cpu_percent"`
	MemoryMB   float64 `json:"memory_mb"`
}

// Port — PRD §8. Port → PID → process (§10-11).
type Port struct {
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	Address     string `json:"address"`
	PID         int    `json:"pid"`
	ProcessName string `json:"process"`
	Command     string `json:"command,omitempty"`
}

// Runtime — PRD §8, §12.
type Runtime struct {
	Name      string `json:"name"`
	Command   string `json:"command"`
	Version   string `json:"version"`
	Path      string `json:"path"`
	Installed bool   `json:"installed"`
}

// NetworkInterface — PRD §8, §13.
type NetworkInterface struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	MAC     string `json:"mac,omitempty"`
}

// NetworkInfo aggregates interfaces + gateway + DNS.
type NetworkInfo struct {
	Interfaces []NetworkInterface `json:"interfaces"`
	Gateway    string             `json:"gateway"`
	DNS        []string           `json:"dns"`
	Hostname   string             `json:"hostname,omitempty"`
}

// Tool — PRD §14.
type Tool struct {
	Name      string `json:"name"`
	Command   string `json:"command"`
	Version   string `json:"version"`
	Path      string `json:"path"`
	Installed bool   `json:"installed"`
}

// Container — PRD §15.
type Container struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	Status  string `json:"status"`
	State   string `json:"state"`
	Ports   string `json:"ports"`
	Runtime string `json:"runtime"` // docker | podman
}

// ProjectInfo — PRD §16.
type ProjectInfo struct {
	Name           string   `json:"name"`
	Path           string   `json:"path"`
	Languages      []string `json:"languages"`
	Tools          []string `json:"tools"`
	Files          []string `json:"files"`
	PackageManager string   `json:"package_manager,omitempty"`
	Framework      string   `json:"framework,omitempty"`
	GitBranch      string   `json:"git_branch,omitempty"`
}

// SystemInfo — PRD §17.
type SystemInfo struct {
	OS          string  `json:"os"`
	Kernel      string  `json:"kernel"`
	Arch        string  `json:"arch"`
	Platform    string  `json:"platform,omitempty"`
	CPUModel    string  `json:"cpu_model"`
	CPUCores    int     `json:"cpu_cores"`
	MemTotalGB  float64 `json:"mem_total_gb"`
	MemUsedGB   float64 `json:"mem_used_gb"`
	DiskTotalGB float64 `json:"disk_total_gb"`
	DiskUsedGB  float64 `json:"disk_used_gb"`
	UptimeDays  float64 `json:"uptime_days"`
	Uptime      string  `json:"uptime"`
}

// DoctorEntry — PRD §20, §36.
type DoctorEntry struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // ok | warning | missing
	Detail  string `json:"detail"`
	Version string `json:"version,omitempty"`
}

// DoctorReport aggregates environment diagnostics.
type DoctorReport struct {
	Entries  []DoctorEntry `json:"entries"`
	OK       int           `json:"ok"`
	Warnings int           `json:"warnings"`
	Missing  int           `json:"missing"`
}

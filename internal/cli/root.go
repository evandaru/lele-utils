package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"lele-dev/internal/app"
	"lele-dev/internal/config"
	"lele-dev/internal/logger"
	"lele-dev/internal/services/container"
	"lele-dev/internal/services/doctor"
	"lele-dev/internal/services/network"
	"lele-dev/internal/services/packages"
	"lele-dev/internal/services/port"
	"lele-dev/internal/services/process"
	"lele-dev/internal/services/project"
	"lele-dev/internal/services/runtime"
	"lele-dev/internal/services/system"
	"lele-dev/internal/services/tools"

	"github.com/spf13/cobra"
)

var (
	jsonOut bool
	debug   bool
	cfg     config.Config
)

// NewRoot builds the cobra tree (PRD §20 CLI mode + §21 JSON + §28 debug).
func NewRoot(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "lele-dev",
		Short: "Developer Workstation Toolkit — TUI + CLI untuk macOS & Linux",
		Long: `lele-dev adalah toolbox interaktif untuk inspeksi workstation developer:
processes, ports, runtimes, network, dev tools, containers, projects, packages, system.

Tanpa argumen → TUI dashboard. Dengan subcommand → CLI non-interaktif
(cocok untuk shell script, CI, AI agent) dengan flag --json.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cfg = config.Load()
			logger.Init(debug)
			logger.Debugf("cmd=%s args=%v", cmd.Name(), args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.Run(cfg)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().BoolVar(&debug, "debug", false, "tulis debug log ke state dir (default silent)")
	root.Version = version

	root.AddCommand(
		newPortsCmd(),
		newProcessCmd(),
		newRuntimesCmd(),
		newNetworkCmd(),
		newDoctorCmd(),
		newToolsCmd(),
		newSystemCmd(),
		newProjectCmd(),
		newContainersCmd(),
		newPackagesCmd(),
		newCleanupCmd(),
	)
	return root
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func fail(err error) error {
	// PRD §25: jangan panic, pesan ramah.
	fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	return err
}

// ---- ports ----

func newPortsCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "ports [port]",
		Short:   "List listening ports / detail port",
		Example: "  lele-dev ports\n  lele-dev ports 3000\n  lele-dev ports --json",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := port.New()
			if len(args) == 1 {
				n, err := strconv.Atoi(args[0])
				if err != nil {
					return fail(fmt.Errorf("port harus angka, dapat %q", args[0]))
				}
				p, err := svc.Find(n)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to inspect port %d.\n\nReason:\n%s\n", n, err.Error())
					return err
				}
				if jsonOut {
					return printJSON(p)
				}
				fmt.Printf("PORT %d\n\nAddress:   %s\nProtocol:  %s\nPID:       %d\nProcess:   %s\n\nCommand:\n%s\n",
					p.Port, p.Address, p.Protocol, p.PID, p.ProcessName, p.Command)
				return nil
			}
			items, err := svc.List()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to inspect ports.\n\nReason:\n%s\n", err.Error())
				return err
			}
			if jsonOut {
				if items == nil {
					fmt.Println("[]")
					return nil
				}
				return printJSON(items)
			}
			fmt.Printf("%-8s %-10s %-8s %-16s %s\n", "PORT", "PROTOCOL", "PID", "ADDRESS", "PROCESS")
			for _, p := range items {
				fmt.Printf("%-8d %-10s %-8d %-16s %s\n", p.Port, p.Protocol, p.PID, p.Address, p.ProcessName)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&jsonOut, "json", false, "output JSON (untuk script/CI/AI agent)")
	return c
}

// ---- process ----

func newProcessCmd() *cobra.Command {
	var useJSON bool
	c := &cobra.Command{
		Use:     "process [pid]",
		Aliases: []string{"processes", "ps"},
		Short:   "List processes / detail process",
		Example: "  lele-dev process\n  lele-dev process 18231\n  lele-dev process --json",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := process.New()
			if len(args) == 1 {
				pid, err := strconv.Atoi(args[0])
				if err != nil {
					return fail(fmt.Errorf("pid harus angka, dapat %q", args[0]))
				}
				p, err := svc.Find(pid)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to inspect process %d.\n\nReason:\n%s\n", pid, err.Error())
					return err
				}
				if useJSON {
					return printJSON(p)
				}
				fmt.Printf("PID:     %d\nName:    %s\nCommand: %s\nUser:    %s\nCPU:     %.1f%%\nMemory:  %.0f MB\n",
					p.PID, p.Name, p.Command, p.User, p.CPUPercent, p.MemoryMB)
				return nil
			}
			items, err := svc.List()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to inspect processes.\n\nReason:\n%s\n", err.Error())
				return err
			}
			if useJSON {
				return printJSON(items)
			}
			fmt.Printf("%-8s %-18s %-8s %-10s %s\n", "PID", "NAME", "CPU", "MEMORY", "COMMAND")
			for _, p := range items {
				fmt.Printf("%-8d %-18s %6.1f%% %8.0fMB %s\n", p.PID, trunc(p.Name, 18), p.CPUPercent, p.MemoryMB, trunc(p.Command, 60))
			}
			return nil
		},
	}
	c.Flags().BoolVar(&useJSON, "json", false, "output JSON")
	return c
}

// ---- runtimes ----

func newRuntimesCmd() *cobra.Command {
	var useJSON bool
	c := &cobra.Command{
		Use:     "runtimes",
		Short:   "Detect programming language runtimes",
		Example: "  lele-dev runtimes\n  lele-dev runtimes --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			items := runtime.New().DetectAll()
			if useJSON {
				return printJSON(items)
			}
			fmt.Printf("%-12s %-14s %s\n", "NAME", "VERSION", "PATH")
			for _, r := range items {
				ver, path := r.Version, r.Path
				if !r.Installed {
					ver, path = "—", "not installed"
				}
				fmt.Printf("%-12s %-14s %s\n", r.Name, ver, path)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&useJSON, "json", false, "output JSON")
	return c
}

// ---- network ----

func newNetworkCmd() *cobra.Command {
	var useJSON bool
	c := &cobra.Command{
		Use:     "network",
		Short:   "Show network interfaces, gateway, DNS",
		Example: "  lele-dev network\n  lele-dev network --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := network.New().Info()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to inspect network.\n\nReason:\n%s\n", err.Error())
				return err
			}
			if useJSON {
				return printJSON(info)
			}
			fmt.Printf("%-12s %-16s %-8s %s\n", "NAME", "ADDRESS", "STATUS", "TYPE")
			for _, ni := range info.Interfaces {
				fmt.Printf("%-12s %-16s %-8s %s\n", ni.Name, ni.Address, ni.Status, ni.Type)
			}
			fmt.Printf("\nGateway\n%s\n\nDNS\n", nonEmpty(info.Gateway, "(unknown)"))
			for _, d := range info.DNS {
				fmt.Println(d)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&useJSON, "json", false, "output JSON")
	return c
}

// ---- doctor ----

func newDoctorCmd() *cobra.Command {
	var useJSON bool
	c := &cobra.Command{
		Use:     "doctor",
		Short:   "Diagnostic developer environment",
		Example: "  lele-dev doctor\n  lele-dev doctor --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			report := doctor.New().Check()
			if useJSON {
				return printJSON(report)
			}
			fmt.Println("LELE-DEV DOCTOR")
			for _, e := range report.Entries {
				icon := "✗"
				switch e.Status {
				case "ok":
					icon = "✓"
				case "warning":
					icon = "⚠"
				}
				if e.Version != "" {
					fmt.Printf("%s %s %s (%s)\n", icon, e.Name, e.Version, e.Detail)
				} else {
					fmt.Printf("%s %s (%s)\n", icon, e.Name, e.Detail)
				}
			}
			fmt.Printf("\n%d OK\n%d Warning\n%d Missing\n", report.OK, report.Warnings, report.Missing)
			return nil
		},
	}
	c.Flags().BoolVar(&useJSON, "json", false, "output JSON")
	return c
}

// ---- tools (v0.2) ----

func newToolsCmd() *cobra.Command {
	var useJSON bool
	c := &cobra.Command{
		Use:   "tools",
		Short: "Detect installed developer tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			items := tools.New().DetectAll()
			if useJSON {
				return printJSON(items)
			}
			fmt.Printf("%-12s %-14s %-6s %s\n", "TOOL", "VERSION", "STATUS", "PATH")
			for _, t := range items {
				st, ver, path := "✗", "—", "not installed"
				if t.Installed {
					st, ver, path = "✓", t.Version, t.Path
				}
				fmt.Printf("%-12s %-14s %-6s %s\n", t.Name, ver, st, path)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&useJSON, "json", false, "output JSON")
	return c
}

// ---- system (v0.2) ----

func newSystemCmd() *cobra.Command {
	var useJSON bool
	c := &cobra.Command{
		Use:   "system",
		Short: "Show CPU / RAM / disk / OS info",
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := system.New().Info()
			if err != nil {
				return fail(err)
			}
			if useJSON {
				return printJSON(info)
			}
			fmt.Printf("OS\n%s (%s)\n\nKernel\n%s\n\nCPU\n%s (%d cores)\n\nMemory\n%.1f GB total, %.1f GB used\n\nDisk\n%.1f GB total, %.1f GB used\n\nUptime\n%s\n",
				info.OS, info.Arch, info.Kernel, info.CPUModel, info.CPUCores,
				info.MemTotalGB, info.MemUsedGB, info.DiskTotalGB, info.DiskUsedGB, info.Uptime)
			return nil
		},
	}
	c.Flags().BoolVar(&useJSON, "json", false, "output JSON")
	return c
}

// ---- project (v0.2) ----

func newProjectCmd() *cobra.Command {
	var useJSON bool
	c := &cobra.Command{
		Use:   "project [dir]",
		Short: "Detect current project type",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := ""
			if len(args) > 0 {
				dir = args[0]
			}
			info, err := project.New().Inspect(dir)
			if err != nil {
				return fail(err)
			}
			if useJSON {
				return printJSON(info)
			}
			fmt.Printf("PROJECT\n\nName:\n%s\n\nPath:\n%s\n", info.Name, info.Path)
			if len(info.Languages) > 0 {
				fmt.Println("\nDetected:")
				for _, l := range info.Languages {
					fmt.Printf("✓ %s\n", l)
				}
			}
			if info.PackageManager != "" {
				fmt.Printf("\nPackage manager: %s\n", info.PackageManager)
			}
			if info.Framework != "" {
				fmt.Printf("Framework: %s\n", info.Framework)
			}
			if len(info.Files) > 0 {
				fmt.Println("\nFiles:")
				for _, f := range info.Files {
					fmt.Println(f)
				}
			}
			return nil
		},
	}
	c.Flags().BoolVar(&useJSON, "json", false, "output JSON")
	return c
}

// ---- containers (v0.3) ----

func newContainersCmd() *cobra.Command {
	var useJSON bool
	c := &cobra.Command{
		Use:     "containers",
		Aliases: []string{"container", "ps-docker"},
		Short:   "List Docker / Podman containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			items, _ := container.New().List()
			if useJSON {
				if items == nil {
					fmt.Println("[]")
					return nil
				}
				return printJSON(items)
			}
			docker, podman := container.Available()
			if len(items) == 0 {
				fmt.Printf("Docker: %s\nPodman: %s\n\n(no containers)\n", availStr(docker), availStr(podman))
				return nil
			}
			fmt.Printf("%-16s %-24s %-10s %-10s %s\n", "NAME", "IMAGE", "RUNTIME", "STATE", "PORTS")
			for _, ct := range items {
				fmt.Printf("%-16s %-24s %-10s %-10s %s\n", ct.Name, ct.Image, ct.Runtime, ct.State, ct.Ports)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&useJSON, "json", false, "output JSON")
	return c
}

// ---- packages ----

func newPackagesCmd() *cobra.Command {
	var useJSON bool
	c := &cobra.Command{
		Use:   "packages [manager]",
		Short: "List installed packages + sizes (npm/pip/brew/pacman/aur/apt/cargo/gem/go)",
		Long: `Menampilkan package terinstall beserta ukurannya dari package manager
yang tersedia: npm (global), pip, brew, pacman (repo), aur (yay/paru),
apt (dpkg), cargo, gem, go (bin). Manager yang tidak terinstall di-skip.`,
		Example: "  lele-dev packages\n  lele-dev packages npm\n  lele-dev packages pacman --json",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := ""
			if len(args) > 0 {
				manager = args[0]
			}
			items, err := packages.New().List(manager)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to inspect packages.\n\nReason:\n%s\n", err.Error())
				return err
			}
			if useJSON {
				if items == nil {
					fmt.Println("[]")
					return nil
				}
				return printJSON(items)
			}
			if len(items) == 0 {
				fmt.Println("(no packages detected — package manager tidak terinstall?)")
				return nil
			}
			fmt.Printf("%-10s %-28s %-16s %s\n", "MANAGER", "NAME", "VERSION", "SIZE")
			for _, p := range items {
				fmt.Printf("%-10s %-28s %-16s %s\n", p.Manager, trunc(p.Name, 28), trunc(p.Version, 16), p.Size)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&useJSON, "json", false, "output JSON")
	return c
}

// ---- cleanup (future §37, CLI basic) ----

func newCleanupCmd() *cobra.Command {
	var killList string
	var yes bool
	var useJSON bool
	c := &cobra.Command{
		Use:   "cleanup",
		Short: "List dev ports for cleanup / kill selected",
		Long: `Menampilkan kandidat development ports. Tidak pernah cleanup otomatis
tanpa confirmation (PRD §37). Gunakan --kill 3000,5173 --yes untuk kill eksplisit.`,
		Example: "  lele-dev cleanup\n  lele-dev cleanup --kill 3000,5173 --yes",
		RunE: func(cmd *cobra.Command, args []string) error {
			ports, err := port.New().List()
			if err != nil {
				return fail(err)
			}
			if killList == "" {
				if useJSON {
					return printJSON(ports)
				}
				fmt.Println("DEV PORT CLEANUP (candidates)")
				for _, p := range ports {
					fmt.Printf("%-6d %-12s PID %-6d %s\n", p.Port, p.ProcessName, p.PID, p.Address)
				}
				fmt.Println("\n[info] gunakan --kill PORT,... --yes untuk kill eksplisit.")
				return nil
			}
			if !yes {
				return fail(fmt.Errorf("cleanup membutuhkan --yes sebagai confirmation eksplisit"))
			}
			svc := process.New()
			for _, s := range strings.Split(killList, ",") {
				s = strings.TrimSpace(s)
				n, err := strconv.Atoi(s)
				if err != nil {
					continue
				}
				p, err := port.New().Find(n)
				if err != nil {
					fmt.Fprintf(os.Stderr, "skip port %d: %s\n", n, err.Error())
					continue
				}
				if err := svc.Kill(p.PID); err != nil {
					fmt.Fprintf(os.Stderr, "gagal kill %d (PID %d): %s\n", n, p.PID, err.Error())
					continue
				}
				fmt.Printf("killed port %d (PID %d %s)\n", n, p.PID, p.ProcessName)
			}
			return nil
		},
	}
	c.Flags().StringVar(&killList, "kill", "", "daftar port dipisah koma, mis. 3000,5173")
	c.Flags().BoolVar(&yes, "yes", false, "confirmation eksplisit (wajib untuk kill)")
	c.Flags().BoolVar(&useJSON, "json", false, "output JSON")
	return c
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func nonEmpty(s, fb string) string {
	if s == "" {
		return fb
	}
	return s
}

func availStr(b bool) string {
	if b {
		return "Installed"
	}
	return "Not installed"
}

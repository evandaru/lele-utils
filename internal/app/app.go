package app

import (
	"fmt"
	"os"
	"os/exec"
	stdruntime "runtime"
	"strings"

	"lele-dev/internal/config"
	"lele-dev/internal/models"
	"lele-dev/internal/services/container"
	"lele-dev/internal/services/doctor"
	"lele-dev/internal/services/network"
	"lele-dev/internal/services/packages"
	"lele-dev/internal/services/port"
	"lele-dev/internal/services/process"
	"lele-dev/internal/services/project"
	rtservice "lele-dev/internal/services/runtime"
	"lele-dev/internal/services/system"
	"lele-dev/internal/services/tools"
	"lele-dev/internal/ui/shared"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Screen identifies current view (PRD §9 router).
type Screen int

const (
	ScreenDashboard Screen = iota
	ScreenProcesses
	ScreenProcessDetail
	ScreenPorts
	ScreenPortDetail
	ScreenRuntimes
	ScreenNetwork
	ScreenTools
	ScreenContainers
	ScreenProjects
	ScreenSystem
	ScreenPackages
)

func (s Screen) String() string {
	switch s {
	case ScreenProcesses:
		return "Processes"
	case ScreenProcessDetail:
		return "Process Detail"
	case ScreenPorts:
		return "Ports"
	case ScreenPortDetail:
		return "Port Detail"
	case ScreenRuntimes:
		return "Runtimes"
	case ScreenNetwork:
		return "Network"
	case ScreenTools:
		return "Dev Tools"
	case ScreenContainers:
		return "Containers"
	case ScreenProjects:
		return "Projects"
	case ScreenSystem:
		return "System"
	case ScreenPackages:
		return "Packages"
	default:
		return "Dashboard"
	}
}

// Messages (async data loading, PRD §39 concurrent scan).
type processesMsg struct {
	items []models.Process
	err   error
}
type portsMsg struct {
	items []models.Port
	err   error
}
type runtimesMsg struct{ items []models.Runtime }
type networkMsg struct {
	info *models.NetworkInfo
	err  error
}
type toolsMsg struct{ items []models.Tool }
type containersMsg struct{ items []models.Container }
type systemMsg struct {
	info *models.SystemInfo
	err  error
}
type projectMsg struct {
	info *models.ProjectInfo
	err  error
}
type packagesMsg struct {
	items []models.Package
}
type killResultMsg struct {
	msg string
	err error
}
type containerActionMsg struct {
	msg string
	err error
}

// Model is the root Bubble Tea model.
type Model struct {
	cfg       config.Config
	screen    Screen
	prev      Screen
	cursor    int
	searching bool
	search    textinput.Model
	query     string

	processes  []models.Process
	ports      []models.Port
	runtimes   []models.Runtime
	netinfo    *models.NetworkInfo
	tools      []models.Tool
	containers []models.Container
	sysinfo    *models.SystemInfo
	project    *models.ProjectInfo
	packages   []models.Package
	pkgsLoaded bool

	detailIdx int // index into current filtered list
	confirm   *confirmState
	message   string
	errMsg    string
	loading   bool
	showHelp  bool
	logsView  string // container logs overlay

	procSvc *process.Service
	portSvc *port.Service
	rtSvc   *rtservice.Service
	netSvc  *network.Service
	toolSvc *tools.Service
	pkgSvc  *packages.Service
	contSvc *container.Service
	projSvc *project.Service
	sysSvc  *system.Service
	docSvc  *doctor.Service
}

type confirmState struct {
	kind   string // "process" | "port-process" | "force"
	pid    int
	label  string
	stage  int // 1 = SIGTERM confirm, 2 = SIGKILL second confirm
	portNo int
}

// New builds the root model (UI owns no system logic — only services).
func New(cfg config.Config) Model {
	ti := textinput.New()
	ti.Prompt = "/ "
	ti.Placeholder = "type to filter, enter to apply, esc to cancel"
	ti.CharLimit = 64
	m := Model{
		cfg:     cfg,
		screen:  ScreenDashboard,
		procSvc: process.New(),
		portSvc: port.New(),
		rtSvc:   rtservice.New(),
		netSvc:  network.New(),
		toolSvc: tools.New(),
		contSvc: container.New(),
		projSvc: project.New(),
		sysSvc:  system.New(),
		docSvc:  doctor.New(),
		pkgSvc:  packages.New(),
		search:  ti,
		loading: true,
	}
	switch strings.ToLower(cfg.DefaultScreen) {
	case "processes":
		m.screen = ScreenProcesses
	case "ports":
		m.screen = ScreenPorts
	case "runtimes":
		m.screen = ScreenRuntimes
	case "network":
		m.screen = ScreenNetwork
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadProcesses(m.procSvc),
		loadPorts(m.portSvc),
		loadRuntimes(m.rtSvc),
		loadNetwork(m.netSvc),
		loadTools(m.toolSvc),
		loadContainers(m.contSvc),
		loadSystem(m.sysSvc),
		loadProject(m.projSvc),
		loadPackages(m.pkgSvc),
	)
}

// ---- loaders ----

func loadProcesses(s *process.Service) tea.Cmd {
	return func() tea.Msg {
		items, err := s.List()
		return processesMsg{items, err}
	}
}
func loadPorts(s *port.Service) tea.Cmd {
	return func() tea.Msg {
		items, err := s.List()
		return portsMsg{items, err}
	}
}
func loadRuntimes(s *rtservice.Service) tea.Cmd {
	return func() tea.Msg { return runtimesMsg{s.DetectAll()} }
}
func loadNetwork(s *network.Service) tea.Cmd {
	return func() tea.Msg {
		info, err := s.Info()
		return networkMsg{info, err}
	}
}
func loadTools(s *tools.Service) tea.Cmd {
	return func() tea.Msg { return toolsMsg{s.DetectAll()} }
}
func loadContainers(s *container.Service) tea.Cmd {
	return func() tea.Msg {
		items, _ := s.List()
		return containersMsg{items}
	}
}
func loadSystem(s *system.Service) tea.Cmd {
	return func() tea.Msg {
		info, err := s.Info()
		return systemMsg{info, err}
	}
}
func loadProject(s *project.Service) tea.Cmd {
	return func() tea.Msg {
		info, err := s.Inspect("")
		return projectMsg{info, err}
	}
}
func loadPackages(s *packages.Service) tea.Cmd {
	return func() tea.Msg {
		items, _ := s.List("")
		if items == nil {
			items = []models.Package{}
		}
		return packagesMsg{items}
	}
}

// ---- Update ----

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.searching {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.query = m.search.Value()
				m.searching = false
				m.cursor = 0
				return m, nil
			case "esc":
				m.searching = false
				m.search.Blur()
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil
	case processesMsg:
		if msg.err == nil {
			m.processes = msg.items
		} else {
			m.errMsg = friendlyErr("process list", msg.err)
		}
		m.loading = false
		return m, nil
	case portsMsg:
		if msg.err == nil {
			m.ports = msg.items
		} else {
			m.errMsg = friendlyErr("port list", msg.err)
		}
		m.loading = false
		return m, nil
	case runtimesMsg:
		m.runtimes = msg.items
		m.loading = false
		return m, nil
	case networkMsg:
		if msg.err == nil {
			m.netinfo = msg.info
		} else {
			m.errMsg = friendlyErr("network info", msg.err)
		}
		m.loading = false
		return m, nil
	case toolsMsg:
		m.tools = msg.items
		m.loading = false
		return m, nil
	case containersMsg:
		m.containers = msg.items
		m.loading = false
		return m, nil
	case systemMsg:
		if msg.err == nil {
			m.sysinfo = msg.info
		}
		m.loading = false
		return m, nil
	case projectMsg:
		if msg.err == nil {
			m.project = msg.info
		}
		return m, nil
	case packagesMsg:
		m.packages = msg.items
		m.pkgsLoaded = true
		m.loading = false
		return m, nil
	case killResultMsg:
		m.confirm = nil
		if msg.err != nil {
			m.errMsg = friendlyErr("kill process", msg.err)
			m.message = ""
		} else {
			m.message = msg.msg
			m.errMsg = ""
		}
		// refresh lists after kill
		return m, tea.Batch(loadProcesses(m.procSvc), loadPorts(m.portSvc))
	case containerActionMsg:
		if msg.err != nil {
			m.errMsg = friendlyErr("container action", msg.err)
		} else {
			m.message = msg.msg
			m.errMsg = ""
		}
		return m, loadContainers(m.contSvc)
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func friendlyErr(what string, err error) string {
	return fmt.Sprintf("Unable to inspect %s.\n\nReason:\n%s\n\nPress Enter to continue.", what, err.Error())
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Error overlay: Enter dismisses (PRD §25).
	if m.errMsg != "" {
		if key == "enter" || key == "esc" {
			m.errMsg = ""
		}
		return m, nil
	}
	// Logs overlay
	if m.logsView != "" {
		m.logsView = ""
		return m, nil
	}
	// Kill confirmation (PRD §10: y/n, force needs 2nd confirm §23.5).
	if m.confirm != nil {
		switch key {
		case "y", "Y":
			c := m.confirm
			if c.stage == 2 {
				pid := c.pid
				m.confirm = nil
				return m, func() tea.Msg {
					err := m.procSvc.KillForce(pid)
					if err != nil {
						return killResultMsg{"", err}
					}
					return killResultMsg{fmt.Sprintf("PID %d force-killed (SIGKILL).", pid), nil}
				}
			}
			pid := c.pid
			svc := m.procSvc
			return m, func() tea.Msg {
				if err := svc.Kill(pid); err != nil {
					// jika SIGTERM gagal karena permission/process pergi, kembalikan error
					return killResultMsg{"", err}
				}
				return killResultMsg{fmt.Sprintf("PID %d terminated (SIGTERM).", pid), nil}
			}
		case "n", "N", "esc":
			m.confirm = nil
			return m, nil
		case "K": // uppercase K pada confirm stage 1 → lanjut ke force confirm
			if m.confirm.stage == 1 {
				m.confirm.stage = 2
			}
			return m, nil
		}
		return m, nil
	}

	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "?":
		m.showHelp = !m.showHelp
		return m, nil
	case "/":
		m.searching = true
		m.search.Focus()
		m.search.SetValue(m.query)
		return m, nil
	case "r", "R":
		m.message = "Refreshing..."
		return m, m.refreshCurrent()
	case "esc":
		m.query = ""
		m.cursor = 0
		if m.screen == ScreenProcessDetail {
			m.screen = ScreenProcesses
			return m, nil
		}
		if m.screen == ScreenPortDetail {
			m.screen = ScreenPorts
			return m, nil
		}
		if m.screen != ScreenDashboard {
			m.screen = ScreenDashboard
			m.cursor = 0
			return m, nil
		}
		return m, nil
	case "up", "k":
		// 'k' is kill on list screens — handled per-screen below.
		// Untuk navigasi gunakan ↑; di sini 'k' hanya navigasi bila bukan kill-context.
		if key == "up" {
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		}
	case "down", "j":
		if key == "down" {
			m.cursor++
			return m, nil
		}
	case "enter":
		return m.enter()
	}

	// Per-screen keys
	switch m.screen {
	case ScreenDashboard:
		return m.dashboardKey(key)
	case ScreenProcesses:
		return m.processesKey(key)
	case ScreenProcessDetail:
		return m.processDetailKey(key)
	case ScreenPorts:
		return m.portsKey(key)
	case ScreenPortDetail:
		return m.portDetailKey(key)
	case ScreenContainers:
		return m.containersKey(key)
	case ScreenPackages:
		return m.packagesKey(key)
	}
	return m, nil
}

func (m Model) refreshCurrent() tea.Cmd {
	switch m.screen {
	case ScreenProcesses, ScreenProcessDetail:
		return loadProcesses(m.procSvc)
	case ScreenPorts, ScreenPortDetail:
		return loadPorts(m.portSvc)
	case ScreenRuntimes:
		return loadRuntimes(m.rtSvc)
	case ScreenNetwork:
		return loadNetwork(m.netSvc)
	case ScreenTools:
		return loadTools(m.toolSvc)
	case ScreenContainers:
		return loadContainers(m.contSvc)
	case ScreenSystem:
		return loadSystem(m.sysSvc)
	case ScreenProjects:
		return loadProject(m.projSvc)
	case ScreenPackages:
		return loadPackages(m.pkgSvc)
	default:
		return tea.Batch(
			loadProcesses(m.procSvc), loadPorts(m.portSvc),
			loadRuntimes(m.rtSvc), loadNetwork(m.netSvc),
			loadSystem(m.sysSvc),
		)
	}
}

func (m Model) enter() (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenDashboard:
		switch m.cursor {
		case 0:
			m.screen = ScreenProcesses
		case 1:
			m.screen = ScreenPorts
		case 2:
			m.screen = ScreenRuntimes
		case 3:
			m.screen = ScreenNetwork
		case 4:
			m.screen = ScreenTools
		case 5:
			m.screen = ScreenContainers
		case 6:
			m.screen = ScreenProjects
		case 7:
			m.screen = ScreenSystem
		case 8:
			m.screen = ScreenPackages
		}
		m.cursor = 0
		m.query = ""
		return m, nil
	case ScreenProcesses:
		items := m.filteredProcesses()
		if len(items) > 0 && m.cursor < len(items) {
			m.detailIdx = items[m.cursor]
			m.prev = ScreenProcesses
			m.screen = ScreenProcessDetail
		}
		return m, nil
	case ScreenPorts:
		items := m.filteredPorts()
		if len(items) > 0 && m.cursor < len(items) {
			m.detailIdx = items[m.cursor]
			m.prev = ScreenPorts
			m.screen = ScreenPortDetail
		}
		return m, nil
	}
	// dismiss error overlay
	if m.errMsg != "" {
		m.errMsg = ""
	}
	return m, nil
}

func (m Model) dashboardKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "1":
		m.screen = ScreenProcesses
	case "2":
		m.screen = ScreenPorts
	case "3":
		m.screen = ScreenRuntimes
	case "4":
		m.screen = ScreenNetwork
	case "5":
		m.screen = ScreenTools
	case "6":
		m.screen = ScreenContainers
	case "7":
		m.screen = ScreenProjects
	case "8":
		m.screen = ScreenSystem
	case "9":
		m.screen = ScreenPackages
	default:
		return m, nil
	}
	m.cursor = 0
	m.query = ""
	return m, nil
}

// ---- filtered helpers ----

func (m *Model) filteredProcesses() []int {
	var rows []string
	for _, p := range m.processes {
		rows = append(rows, fmt.Sprintf("%d %s %s", p.PID, p.Name, p.Command))
	}
	return shared.Filter(rows, m.query)
}

func (m *Model) filteredPorts() []int {
	var rows []string
	for _, p := range m.ports {
		rows = append(rows, fmt.Sprintf("%d %s %s %s", p.Port, p.ProcessName, p.Address, p.Protocol))
	}
	return shared.Filter(rows, m.query)
}

func (m *Model) filteredPackages() []int {
	var rows []string
	for _, p := range m.packages {
		rows = append(rows, fmt.Sprintf("%s %s %s", p.Manager, p.Name, p.Version))
	}
	return shared.Filter(rows, m.query)
}

func (m *Model) filteredContainers() []int {
	var rows []string
	for _, c := range m.containers {
		rows = append(rows, fmt.Sprintf("%s %s %s %s", c.Name, c.Image, c.Status, c.Runtime))
	}
	return shared.Filter(rows, m.query)
}

// ---- per-screen keys ----

func (m Model) processesKey(key string) (tea.Model, tea.Cmd) {
	items := m.filteredProcesses()
	clamp := func() {
		if len(items) == 0 {
			m.cursor = 0
			return
		}
		if m.cursor >= len(items) {
			m.cursor = len(items) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
	}
	switch key {
	case "down", "j":
		m.cursor++
		clamp()
		return m, nil
	case "up", "k":
		// 'k' di process list = kill (PRD). Bedakan: jika query/search kosong dan user tekan k → kill.
		// Untuk navigasi atas gunakan ↑. Di sini treat "k" sebagai kill bila ada item.
		if key == "k" {
			if len(items) == 0 {
				return m, nil
			}
			clamp()
			p := m.processes[items[m.cursor]]
			if m.cfg.ConfirmKill {
				m.confirm = &confirmState{kind: "process", pid: p.PID, label: fmt.Sprintf("%d (%s)", p.PID, p.Name), stage: 1}
			} else {
				pid := p.PID
				return m, func() tea.Msg {
					if err := m.procSvc.Kill(pid); err != nil {
						return killResultMsg{"", err}
					}
					return killResultMsg{fmt.Sprintf("PID %d terminated.", pid), nil}
				}
			}
			return m, nil
		}
		m.cursor--
		clamp()
		return m, nil
	}
	return m, nil
}

func (m Model) processDetailKey(key string) (tea.Model, tea.Cmd) {
	if m.detailIdx < 0 || m.detailIdx >= len(m.processes) {
		m.screen = ScreenProcesses
		return m, nil
	}
	p := m.processes[m.detailIdx]
	switch key {
	case "k":
		if m.cfg.ConfirmKill {
			m.confirm = &confirmState{kind: "process", pid: p.PID, label: fmt.Sprintf("%d (%s)", p.PID, p.Name), stage: 1}
		} else {
			pid := p.PID
			return m, func() tea.Msg {
				if err := m.procSvc.Kill(pid); err != nil {
					return killResultMsg{"", err}
				}
				return killResultMsg{fmt.Sprintf("PID %d terminated.", pid), nil}
			}
		}
	}
	return m, nil
}

func (m Model) portsKey(key string) (tea.Model, tea.Cmd) {
	items := m.filteredPorts()
	clamp := func() {
		if len(items) == 0 {
			m.cursor = 0
			return
		}
		if m.cursor >= len(items) {
			m.cursor = len(items) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
	}
	switch key {
	case "down", "j":
		m.cursor++
		clamp()
		return m, nil
	case "up":
		m.cursor--
		clamp()
		return m, nil
	case "k":
		if len(items) == 0 {
			return m, nil
		}
		clamp()
		p := m.ports[items[m.cursor]]
		if p.PID <= 0 {
			m.errMsg = friendlyErr(fmt.Sprintf("port %d", p.Port), fmt.Errorf("PID tidak diketahui (butuh permission lebih tinggi)"))
			return m, nil
		}
		m.confirm = &confirmState{kind: "port-process", pid: p.PID, label: fmt.Sprintf("port %d → PID %d (%s)", p.Port, p.PID, p.ProcessName), stage: 1, portNo: p.Port}
		return m, nil
	case "o": // open localhost (PRD §11)
		if len(items) == 0 {
			return m, nil
		}
		clamp()
		p := m.ports[items[m.cursor]]
		openBrowser(fmt.Sprintf("http://localhost:%d", p.Port))
		m.message = fmt.Sprintf("Opening http://localhost:%d ...", p.Port)
		return m, nil
	case "c": // copy port number
		if len(items) == 0 {
			return m, nil
		}
		clamp()
		p := m.ports[items[m.cursor]]
		if copyToClipboard(fmt.Sprint(p.Port)) {
			m.message = fmt.Sprintf("Port %d copied.", p.Port)
		} else {
			m.message = fmt.Sprintf("Port: %d", p.Port)
		}
		return m, nil
	}
	return m, nil
}

func (m Model) portDetailKey(key string) (tea.Model, tea.Cmd) {
	if m.detailIdx < 0 || m.detailIdx >= len(m.ports) {
		m.screen = ScreenPorts
		return m, nil
	}
	p := m.ports[m.detailIdx]
	switch key {
	case "k":
		if p.PID <= 0 {
			m.errMsg = friendlyErr(fmt.Sprintf("port %d", p.Port), fmt.Errorf("PID tidak diketahui"))
			return m, nil
		}
		m.confirm = &confirmState{kind: "port-process", pid: p.PID, label: fmt.Sprintf("port %d → PID %d (%s)", p.Port, p.PID, p.ProcessName), stage: 1, portNo: p.Port}
	case "o":
		openBrowser(fmt.Sprintf("http://localhost:%d", p.Port))
		m.message = fmt.Sprintf("Opening http://localhost:%d ...", p.Port)
	}
	return m, nil
}

func (m Model) containersKey(key string) (tea.Model, tea.Cmd) {
	items := m.filteredContainers()
	if len(items) == 0 {
		return m, nil
	}
	if m.cursor >= len(items) {
		m.cursor = len(items) - 1
	}
	c := m.containers[items[m.cursor]]
	svc := m.contSvc
	rt, id := c.Runtime, c.ID
	name := c.Name
	switch key {
	case "down", "j":
		if m.cursor < len(items)-1 {
			m.cursor++
		}
		return m, nil
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case "s":
		return m, func() tea.Msg {
			if err := svc.Start(rt, id); err != nil {
				return containerActionMsg{"", err}
			}
			return containerActionMsg{fmt.Sprintf("%s started.", name), nil}
		}
	case "t":
		return m, func() tea.Msg {
			if err := svc.Stop(rt, id); err != nil {
				return containerActionMsg{"", err}
			}
			return containerActionMsg{fmt.Sprintf("%s stopped.", name), nil}
		}
	case "r":
		return m, func() tea.Msg {
			if err := svc.Restart(rt, id); err != nil {
				return containerActionMsg{"", err}
			}
			return containerActionMsg{fmt.Sprintf("%s restarted.", name), nil}
		}
	case "l":
		out, err := svc.Logs(rt, id, 50)
		if err != nil {
			m.errMsg = friendlyErr("container logs", err)
			return m, nil
		}
		m.logsView = fmt.Sprintf("LOGS %s (%s)\n%s", name, rt, out)
		return m, nil
	case "k":
		return m, func() tea.Msg {
			if err := svc.Kill(rt, id); err != nil {
				return containerActionMsg{"", err}
			}
			return containerActionMsg{fmt.Sprintf("%s killed.", name), nil}
		}
	}
	return m, nil
}

// packagesKey — list read-only: navigasi + search + refresh (global).
func (m Model) packagesKey(key string) (tea.Model, tea.Cmd) {
	items := m.filteredPackages()
	switch key {
	case "down", "j":
		if m.cursor < len(items)-1 {
			m.cursor++
		}
		return m, nil
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	}
	return m, nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch stdruntime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func copyToClipboard(s string) bool {
	candidates := [][]string{
		{"pbcopy"},
		{"wl-copy"},
		{"xclip", "-selection", "clipboard"},
		{"xsel", "--clipboard", "--input"},
	}
	for _, c := range candidates {
		if _, err := exec.LookPath(c[0]); err != nil {
			continue
		}
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Stdin = strings.NewReader(s)
		if cmd.Run() == nil {
			return true
		}
	}
	return false
}

// RootWarning — PRD §23.1.
func rootWarning() string {
	if os.Geteuid() == 0 {
		return shared.Yellow.Render("⚠ running as root — tidak disarankan. Fitur normal tidak membutuhkan root.")
	}
	return ""
}

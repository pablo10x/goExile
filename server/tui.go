package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"exile/server/auth"
	"exile/server/database"
	"exile/server/enrollment"
	"exile/server/redeye"
	"exile/server/registry"
	"exile/server/sse"
	"exile/server/utils"
	"exile/server/ws"
	"exile/server/ws_player"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type StartupState struct {
	AuthConfig   auth.AuthConfig
	SessionStore *auth.SessionStore
	SSEHub       *sse.SSEHub
	Router       *mux.Router
}

var GlobalStartup StartupState

type status int

const (
	statusPending status = iota
	statusLoading
	statusDone
	statusFailed
)

type step struct {
	name    string
	status  status
	message string
	action  func(*tuiModel) error
}

type appState int

const (
	stateCheckConfig appState = iota
	stateConfigMenu
	stateConfigPort
	stateConfigDB
	stateRunningSteps
	stateDone
)

type tuiModel struct {
	state          appState
	cursor         int
	textInput      textinput.Model
	steps          []step
	current        int
	spinner        spinner.Model
	progress       progress.Model
	quitting       bool
	err            error
	done           bool
	startTime      time.Time
	GeneratedCreds map[string]string

	// Auto-select
	autoSelectLeft int
	userInteracted bool

	// Config choices
	selectedDriver string
	selectedPort   string
}

const asciiLogo = `
 ███████╗██╗  ██╗██╗██╗     ███████╗
 ██╔════╝╚██╗██╔╝██║██║     ██╔════╝
 █████╗   ╚███╔╝ ██║██║     █████╗  
 ██╔══╝   ██╔██╗ ██║██║     ██╔══╝  
 ███████╗██╔╝ ██╗██║███████╗███████╗
 ╚══════╝╚═╝  ╚═╝╚═╝╚══════╝╚══════╝`

var (
	// Cyberpunk Palette
	colorNeonOrange = lipgloss.Color("#FF5F00")
	colorNeonGreen  = lipgloss.Color("#00FF00")
	colorNeonBlue   = lipgloss.Color("#00FFFF")
	colorNeonPink   = lipgloss.Color("#FF00FF")
	colorDarkGray   = lipgloss.Color("#1a1a1a")
	colorLightGray  = lipgloss.Color("#767676")
	colorWhite      = lipgloss.Color("#FFFFFF")

	// Styles
	appStyle = lipgloss.NewStyle().
			Margin(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorNeonBlue).
			Padding(1, 2)

	logoStyle = lipgloss.NewStyle().
			Foreground(colorNeonOrange).
			Bold(true).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorNeonBlue).
			Border(lipgloss.DoubleBorder(), false, false, true, false).
			BorderForeground(colorNeonPink).
			Padding(0, 1).
			MarginBottom(1)

	footerStyle = lipgloss.NewStyle().
			Foreground(colorLightGray).
			MarginTop(1).
			Italic(true)

	// List Styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(colorNeonGreen).
				Bold(true).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(colorNeonPink).
				PaddingLeft(1)

	itemStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			PaddingLeft(2)

	// Step Styles
	doneStyle = lipgloss.NewStyle().
			Foreground(colorNeonGreen).
			SetString("✔")

	pendingStyle = lipgloss.NewStyle().
			Foreground(colorLightGray).
			SetString("○")

	loadingStyle = lipgloss.NewStyle().
			Foreground(colorNeonBlue)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorNeonPink)

	subItemStyle = lipgloss.NewStyle().
			Foreground(colorLightGray).
			MarginLeft(4)
)

func newTUIModel() tuiModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorNeonOrange)

	ti := textinput.New()
	ti.Placeholder = "8081"
	ti.CharLimit = 5
	ti.Width = 20

	return tuiModel{
		state:     stateCheckConfig,
		textInput: ti,
		steps: []step{
			{name: "Load Configuration", action: loadConfigStep},
			{name: "Security & Keys Check", action: ensureKeysStep},
			{name: "Initialize Authentication", action: initAuthStep},
			{name: "Setup Database Engine", action: setupDBEngineStep},
			{name: "Connect to Database", action: connectDBStep},
			{name: "Initialize Registry", action: initRegistryStep},
			{name: "Start Background Services", action: startServicesStep},
		},
		spinner:        s,
		progress:       progress.New(progress.WithDefaultGradient()),
		startTime:      time.Now(),
		GeneratedCreds: make(map[string]string),
	}
}

func (m tuiModel) Init() tea.Cmd {
	return checkConfigCmd
}

func checkConfigCmd() tea.Msg {
	// Check if key config exists to decide whether to auto-start or show menu
	if os.Getenv("DB_DRIVER") == "" || os.Getenv("MASTER_API_KEY") == "" {
		return stateConfigMenu // Force menu if not configured
	}
	return stateConfigMenu
}

type autoSelectTickMsg time.Time

func autoSelectTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return autoSelectTickMsg(t)
	})
}

type stepDoneMsg struct {
	index int
	err   error
	msg   string
}

func nextStepCmd(index int) tea.Cmd {
	return func() tea.Msg {
		return stepDoneMsg{index: -1} // Trigger first step
	}
}

func runStepCmd(m *tuiModel, index int) tea.Cmd {
	return func() tea.Msg {
		m.steps[index].status = statusLoading
		err := m.steps[index].action(m)
		msg := m.steps[index].message
		if err != nil {
			return stepDoneMsg{index: index, err: err}
		}
		return stepDoneMsg{index: index, msg: msg}
	}
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Cancel auto-select on any key press
		if m.state == stateConfigMenu && !m.userInteracted {
			m.userInteracted = true
		}

		if msg.Type == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}

		// Global q/esc to quit if in menu/config states
		if m.state != stateRunningSteps && (msg.String() == "q" || msg.Type == tea.KeyEsc) {
			m.quitting = true
			return m, tea.Quit
		}

		switch m.state {
		case stateConfigMenu:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < 1 { // 2 options
					m.cursor++
				}
			case "enter":
				if m.cursor == 0 {
					// Start Server (Skip Config)
					m.state = stateRunningSteps
					return m, tea.Batch(m.spinner.Tick, nextStepCmd(0))
				} else {
					// Configure
					m.state = stateConfigPort
					m.textInput.Placeholder = "8081"
					m.textInput.Focus()
					return m, textinput.Blink
				}
			}

		case stateConfigPort:
			switch msg.Type {
			case tea.KeyEnter:
				port := m.textInput.Value()
				if port == "" {
					port = "8081"
				}
				m.selectedPort = port
				_ = updateEnvFile("SERVER_PORT", port) // Save immediately

				m.state = stateConfigDB
				m.cursor = 0 // Reset cursor for next menu
				return m, nil
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd

		case stateConfigDB:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < 1 { // 2 options: Embedded vs External
					m.cursor++
				}
			case "enter":
				if m.cursor == 0 {
					m.selectedDriver = "embedded"
					os.Setenv("DB_DRIVER", "postgres")
					os.Setenv("DB_DSN", "")
					_ = updateEnvFile("DB_DRIVER", "postgres")
					_ = updateEnvFile("DB_DSN", "")
				} else {
					m.selectedDriver = "external"
				}

				m.state = stateRunningSteps
				return m, tea.Batch(m.spinner.Tick, nextStepCmd(0))
			}
		}

	case appState:
		m.state = msg
		// If entering menu, start auto-select timer
		if m.state == stateConfigMenu {
			m.autoSelectLeft = 3
			m.userInteracted = false
			return m, autoSelectTickCmd()
		}
		return m, nil

	case autoSelectTickMsg:
		if m.state == stateConfigMenu && !m.userInteracted {
			m.autoSelectLeft--
			if m.autoSelectLeft <= 0 {
				// Auto-select first option (Start Server)
				m.state = stateRunningSteps
				return m, tea.Batch(m.spinner.Tick, nextStepCmd(0))
			}
			return m, autoSelectTickCmd()
		}

	case spinner.TickMsg:
		if m.state == stateRunningSteps {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case stepDoneMsg:
		if m.state != stateRunningSteps {
			return m, nil
		}

		if msg.index == -1 {
			return m, runStepCmd(&m, 0)
		}

		if msg.err != nil {
			m.steps[msg.index].status = statusFailed
			m.steps[msg.index].message = msg.err.Error()
			m.err = msg.err
			m.quitting = true
			return m, tea.Quit
		}

		m.steps[msg.index].status = statusDone
		m.steps[msg.index].message = msg.msg
		m.current = msg.index + 1

		progressCmd := m.progress.SetPercent(float64(m.current) / float64(len(m.steps)))

		if m.current < len(m.steps) {
			return m, tea.Batch(runStepCmd(&m, m.current), progressCmd)
		}

		m.done = true
		return m, tea.Quit

	case progress.FrameMsg:
		if m.state == stateRunningSteps {
			progressModel, cmd := m.progress.Update(msg)
			m.progress = progressModel.(progress.Model)
			return m, cmd
		}
	}

	return m, nil
}

func (m tuiModel) View() string {
	var content string

	switch m.state {
	case stateCheckConfig:
		content = m.viewCheckConfig()
	case stateConfigMenu:
		content = m.viewConfigMenu()
	case stateConfigPort:
		content = m.viewConfigPort()
	case stateConfigDB:
		content = m.viewConfigDB()
	case stateRunningSteps, stateDone:
		content = m.viewSteps()
	}

	return appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			logoStyle.Render(asciiLogo),
			headerStyle.Render("MASTER SERVER CONTROL"),
			lipgloss.NewStyle().MarginTop(1).Render(content),
			m.viewFooter(),
		),
	)
}

func (m tuiModel) viewCheckConfig() string {
	return fmt.Sprintf("\n%s Checking system configuration...", m.spinner.View())
}

func (m tuiModel) viewConfigMenu() string {
	var s strings.Builder
	s.WriteString("Select an action:\n\n")

	options := []string{"Start Server (Use existing config)", "Configure Server Settings"}
	for i, choice := range options {
		cursor := " "
		style := itemStyle
		if m.cursor == i {
			cursor = "❯"
			style = selectedItemStyle
		}

		// Append countdown to first option if active
		text := choice
		if i == 0 && !m.userInteracted && m.autoSelectLeft > 0 {
			text = fmt.Sprintf("%s (Auto-starting in %ds)", choice, m.autoSelectLeft)
		}

		s.WriteString(style.Render(fmt.Sprintf("%s %s", cursor, text)) + "\n")
	}
	return s.String()
}

func (m tuiModel) viewConfigPort() string {
	var s strings.Builder
	s.WriteString("Configuration > Port\n\n")
	s.WriteString("Enter the port for the Master Server to listen on:\n\n")
	s.WriteString(m.textInput.View() + "\n\n")
	s.WriteString(subItemStyle.Render("(Default: 8081)"))
	return s.String()
}

func (m tuiModel) viewConfigDB() string {
	var s strings.Builder
	s.WriteString("Configuration > Database\n\n")
	s.WriteString("Select Database Driver:\n\n")

	options := []string{"Embedded PostgreSQL (Recommended)", "External PostgreSQL"}
	for i, choice := range options {
		cursor := " "
		style := itemStyle
		if m.cursor == i {
			cursor = "❯"
			style = selectedItemStyle
		}
		s.WriteString(style.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n")
	}
	s.WriteString("\n" + subItemStyle.Render("Embedded: Zero-config, runs locally."))
	s.WriteString("\n" + subItemStyle.Render("External: Requires DB_DSN in .env."))
	return s.String()
}

func (m tuiModel) viewSteps() string {
	var s strings.Builder
	s.WriteString("System Initialization:\n\n")

	for i, step := range m.steps {
		if i < m.current {
			s.WriteString(fmt.Sprintf(" %s %s\n", doneStyle.String(), step.name))
			if step.message != "" {
				s.WriteString(subItemStyle.Render(step.message) + "\n")
			}
		} else if i == m.current && !m.done {
			s.WriteString(fmt.Sprintf(" %s %s\n", m.spinner.View(), loadingStyle.Render(step.name)))
			if step.message != "" {
				s.WriteString(subItemStyle.Render(step.message) + "\n")
			}
		} else {
			s.WriteString(fmt.Sprintf(" %s %s\n", pendingStyle.String(), step.name))
		}
	}

	if m.err != nil {
		s.WriteString("\n" + errorStyle.Render("Error: "+m.err.Error()) + "\n")
	} else if m.done {
		s.WriteString("\n" + doneStyle.Render(fmt.Sprintf("✓ Ready in %v", time.Since(m.startTime).Round(time.Millisecond))) + "\n")
	} else {
		s.WriteString("\n" + m.progress.View() + "\n")
	}

	return s.String()
}

func (m tuiModel) viewFooter() string {
	var help string
	switch m.state {
	case stateConfigMenu, stateConfigDB:
		help = "↑/↓: Select • Enter: Confirm • q: Quit"
	case stateConfigPort:
		help = "Enter: Confirm • q: Quit"
	case stateRunningSteps:
		help = "Initializing..."
	case stateDone:
		help = "Starting Server..."
	}
	return footerStyle.Render(help)
}

func loadConfigStep(m *tuiModel) error {
	_ = godotenv.Load()
	GlobalStartup.Router = mux.NewRouter()
	time.Sleep(100 * time.Millisecond)
	return nil
}

func initAuthStep(m *tuiModel) error {
	GlobalStartup.AuthConfig = auth.GetAuthConfig()
	GlobalStartup.SessionStore = auth.NewSessionStore(GlobalStartup.AuthConfig.IsProduction)
	go GlobalStartup.SessionStore.CleanupExpiredSessions()

	enrollment.InitializeEnrollmentManager()

	GlobalStartup.SSEHub = sse.NewSSEHub()
	go GlobalStartup.SSEHub.Run()

	time.Sleep(100 * time.Millisecond)
	return nil
}

func setupDBEngineStep(m *tuiModel) error {
	driver := os.Getenv("DB_DRIVER")
	dsn := os.Getenv("DB_DSN")

	if driver == "" {
		driver = "postgres"
		os.Setenv("DB_DRIVER", "postgres")
		_ = updateEnvFile("DB_DRIVER", "postgres")
		m.steps[m.current].message = "No driver specified. Defaulting to PostgreSQL."
		time.Sleep(500 * time.Millisecond)
	}

	if driver == "postgres" && dsn == "" {
		m.steps[m.current].message = "Installing Embedded PostgreSQL..."
		if err := database.StartEmbeddedPostgres(); err != nil {
			return err
		}
		defaultDSN := "postgres://exile:exile@localhost:5432/exile_master?sslmode=disable"
		os.Setenv("DB_DSN", defaultDSN)
		_ = updateEnvFile("DB_DSN", defaultDSN)
		m.steps[m.current].message = "Embedded PostgreSQL started."
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

func ensureKeysStep(m *tuiModel) error {
	updates := make(map[string]string)

	if os.Getenv("MASTER_API_KEY") == "" {
		key := utils.GenerateRandomString(32)
		os.Setenv("MASTER_API_KEY", key)
		updates["MASTER_API_KEY"] = key
		m.GeneratedCreds["MASTER_API_KEY"] = key
	}

	if os.Getenv("GAME_API_KEY") == "" {
		key := utils.GenerateRandomString(32)
		os.Setenv("GAME_API_KEY", key)
		updates["GAME_API_KEY"] = key
		m.GeneratedCreds["GAME_API_KEY"] = key
	}

	if os.Getenv("ADMIN_PASSWORD") == "" {
		pass := utils.GenerateRandomString(16)
		os.Setenv("ADMIN_PASSWORD", pass)
		updates["ADMIN_PASSWORD"] = pass
		m.GeneratedCreds["ADMIN_PASSWORD"] = pass
	}

	if len(updates) > 0 {
		for k, v := range updates {
			if err := updateEnvFile(k, v); err != nil {
				return err
			}
		}
		m.steps[m.current].message = fmt.Sprintf("Generated %d missing security keys", len(updates))
	} else {
		m.steps[m.current].message = "Security keys present"
	}

	time.Sleep(200 * time.Millisecond)
	return nil
}

func updateEnvFile(key, value string) error {
	path := ".env"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		f.Close()
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	found := false
	newLines := make([]string, 0, len(lines)+1)

	for _, line := range lines {
		if strings.HasPrefix(line, key+"=") {
			newLines = append(newLines, fmt.Sprintf("%s=%s", key, value))
			found = true
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found {
		if len(newLines) > 0 && newLines[len(newLines)-1] != "" {
			newLines = append(newLines, "")
		}
		newLines = append(newLines, fmt.Sprintf("%s=%s", key, value))
	}

	output := strings.Join(newLines, "\n")
	if !strings.HasSuffix(output, "\n") {
		output += "\n"
	}

	return os.WriteFile(path, []byte(output), 0600)
}

func connectDBStep(m *tuiModel) error {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = os.Getenv("DB_PATH")
		if dsn == "" {
			dsn = "database/registry.db"
		}
	}
	if err := database.InitDB(dsn); err != nil {
		return err
	}
	registry.GlobalStats.SetDBConnected(true)

	roDSN := os.Getenv("READONLY_DB_DSN")
	if roDSN != "" {
		_ = database.InitReadOnlyDB(roDSN)
	} else {
		database.ReadOnlyDBConn = database.DBConn
	}

	time.Sleep(100 * time.Millisecond)
	return nil
}

func initRegistryStep(m *tuiModel) error {
	// Seed default config first
	if err := database.SeedDefaultConfig(database.DBConn); err != nil {
		return fmt.Errorf("seed config: %w", err)
	}

	loaded, err := database.LoadNodes(database.DBConn)
	if err != nil {
		return err
	}

	maxID := registry.GetNextID() - 1
	for i := range loaded {
		s := loaded[i]
		copyS := s
		if copyS.Status == "Online" && !ws.GlobalWSManager.IsClientConnected(copyS.ID) {
			copyS.Status = "Offline"
		}
		registry.SetItem(copyS.ID, &copyS)
		if copyS.ID > maxID {
			maxID = copyS.ID
		}
	}

	if err := database.InitPlayerSystem(database.DBConn); err != nil {
		return err
	}

	registry.GlobalStats.InitializeStats(database.DBConn)
	m.steps[m.current].message = fmt.Sprintf("Restored %d nodes from database", len(loaded))

	time.Sleep(100 * time.Millisecond)
	return nil
}

func startServicesStep(m *tuiModel) error {
	if database.DBConn != nil {
		redeye.StartRedEyeBackground(database.DBConn)
	}

	_ = auth.InitFirebase()

	go ws.GlobalWSManager.Run()
	ws_player.InitPlayerWS()

	time.Sleep(100 * time.Millisecond)
	return nil
}

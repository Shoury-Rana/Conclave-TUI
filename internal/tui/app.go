package tui

import (
	"context"

	"Conclave-TUI/internal/api"
	"Conclave-TUI/internal/config"
	"Conclave-TUI/internal/tui/common"
	"Conclave-TUI/internal/tui/styles"
	"Conclave-TUI/internal/tui/views"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ScreenState enumerates the major views in the TUI application lifecycle.
type ScreenState int

const (
	ScreenLoading ScreenState = iota
	ScreenAuth
	ScreenWorkspace
)

// AppModel is the root Bubble Tea model managing global application state,
// active view routing, responsive resizing, and authentication lifecycle.
type AppModel struct {
	cfg    *config.Config
	client *api.Client
	state  ScreenState

	// Sub-views
	authView       views.AuthModel
	workspaceView  views.WorkspaceModel
	loadingSpinner spinner.Model

	// Viewport dimensions
	width  int
	height int
}

// NewAppModel initializes the Conclave TUI application.
func NewAppModel(cfg *config.Config, client *api.Client) AppModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorSecondary)

	// Set client tokens from config if already present
	if cfg.HasTokens() {
		client.SetTokens(cfg.AccessToken, cfg.RefreshToken)
	}

	authView := views.NewAuthModel(client)
	workspaceView := views.NewWorkspaceModel(cfg, client)

	initialState := ScreenAuth
	if cfg.HasTokens() {
		initialState = ScreenLoading
	}

	return AppModel{
		cfg:            cfg,
		client:         client,
		state:          initialState,
		authView:       authView,
		workspaceView:  workspaceView,
		loadingSpinner: sp,
	}
}

// Init starts the initial Bubble Tea command loop.
// If saved tokens are found in config, it checks their validity against GET /auth/me/.
func (m AppModel) Init() tea.Cmd {
	if m.cfg.HasTokens() {
		return tea.Batch(
			m.loadingSpinner.Tick,
			m.validateSavedTokenCmd(),
		)
	}
	return m.authView.Init()
}

// validateSavedTokenCmd checks if the existing stored access token is valid,
// or refreshes it via the client's built-in refresh mechanism.
func (m AppModel) validateSavedTokenCmd() tea.Cmd {
	return func() tea.Msg {
		me, err := m.client.GetMe(context.Background())
		if err != nil {
			return common.TokenInvalidMsg{Err: err}
		}
		return common.TokenValidatedMsg{Me: me}
	}
}

// Update coordinates incoming messages, global keybindings, and active view updates.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// 1. Terminal Window Resized
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.authView.SetSize(msg.Width, msg.Height)
		m.workspaceView.SetSize(msg.Width, msg.Height)

	// 2. Global Key Press Events
	case tea.KeyPressMsg:
		key := msg.String()

		// Always allow quitting via Ctrl+C
		if key == "ctrl+c" {
			return m, tea.Quit
		}

		// Logout shortcut
		if key == "ctrl+l" && m.state == ScreenWorkspace {
			return m, m.handleLogoutCmd()
		}

		// In workspace view, 'q' also quits if not typing in an input
		if key == "q" && m.state == ScreenWorkspace && m.workspaceView.ActiveTenant() != nil {
			return m, tea.Quit
		}

	// 3. Saved Token Verification Result
	case common.TokenValidatedMsg:
		m.state = ScreenWorkspace
		// Update cached user info in config
		m.cfg.User = config.UserProfile{
			ID:    msg.Me.ID,
			Email: msg.Me.Email,
			Name:  msg.Me.Name,
		}
		_ = m.cfg.Save()

		return m, m.workspaceView.Init()

	case common.TokenInvalidMsg:
		// Saved token is invalid or expired
		_ = m.cfg.ClearAuth()
		m.client.SetTokens("", "")
		m.state = ScreenAuth
		m.authView.SetError(msg.Err)
		return m, m.authView.Init()

	// 4. Successful Authentication (Login or Signup)
	case common.LoginSuccessMsg:
		m.state = ScreenWorkspace
		m.cfg.AccessToken = msg.Response.AccessToken
		m.cfg.RefreshToken = msg.Response.RefreshToken
		m.cfg.User = config.UserProfile{
			ID:    msg.Response.User.ID,
			Email: msg.Response.User.Email,
			Name:  msg.Response.User.Name,
		}
		_ = m.cfg.Save()

		return m, m.workspaceView.Init()

	case common.SignupSuccessMsg:
		m.state = ScreenWorkspace
		m.cfg.AccessToken = msg.Response.AccessToken
		m.cfg.RefreshToken = msg.Response.RefreshToken
		m.cfg.User = config.UserProfile{
			ID:    msg.Response.User.ID,
			Email: msg.Response.User.Email,
			Name:  msg.Response.User.Name,
		}
		_ = m.cfg.Save()

		return m, m.workspaceView.Init()

	// 5. Auth Failure
	case common.AuthErrorMsg:
		m.authView.SetError(msg.Err)
		return m, nil

	// 6. Logout
	case common.LogoutMsg:
		return m, m.handleLogoutCmd()
	}

	// Route events to active sub-view
	switch m.state {
	case ScreenLoading:
		if _, ok := msg.(spinner.TickMsg); ok {
			var cmd tea.Cmd
			m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case ScreenAuth:
		var cmd tea.Cmd
		m.authView, cmd = m.authView.Update(msg)
		cmds = append(cmds, cmd)

	case ScreenWorkspace:
		var cmd tea.Cmd
		m.workspaceView, cmd = m.workspaceView.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *AppModel) handleLogoutCmd() tea.Cmd {
	_ = m.cfg.ClearAuth()
	m.client.SetTokens("", "")
	m.state = ScreenAuth
	m.authView = views.NewAuthModel(m.client)
	m.authView.SetSize(m.width, m.height)
	m.workspaceView.ClearActiveTenant()
	return m.authView.Init()
}

// View returns the declarative terminal view.
// In Bubble Tea v2, we return tea.View with AltScreen enabled.
func (m AppModel) View() tea.View {
	var content string

	switch m.state {
	case ScreenLoading:
		loadingBox := lipgloss.JoinVertical(
			lipgloss.Center,
			styles.LogoStyle.Render(styles.AsciiLogo),
			"\n",
			lipgloss.JoinHorizontal(
				lipgloss.Center,
				m.loadingSpinner.View(),
				" ",
				styles.SubtitleStyle.Render("Verifying Conclave session..."),
			),
		)
		if m.width > 0 && m.height > 0 {
			content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, loadingBox)
		} else {
			content = loadingBox
		}

	case ScreenAuth:
		content = m.authView.View()

	case ScreenWorkspace:
		content = m.workspaceView.View()
	}

	// Declaratively enable AltScreen so the TUI uses a clean full-terminal buffer
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

package views

import (
	"fmt"

	"Conclave-TUI/internal/api"
	"Conclave-TUI/internal/config"
	"Conclave-TUI/internal/tui/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// WorkspaceStubModel serves as the landing screen once the user is authenticated in Phase 1.
// In Phase 2, this will be expanded into the full interactive Workspace Selector.
type WorkspaceStubModel struct {
	cfg  *config.Config
	user *api.User

	width  int
	height int
}

// NewWorkspaceStubModel creates a new workspace dashboard stub.
func NewWorkspaceStubModel(cfg *config.Config) WorkspaceStubModel {
	return WorkspaceStubModel{
		cfg: cfg,
	}
}

// SetUser updates the currently displayed user profile.
func (m *WorkspaceStubModel) SetUser(user *api.User) {
	m.user = user
}

// SetSize updates the viewport dimensions for responsive centering.
func (m *WorkspaceStubModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Init implements tea.Model.
func (m WorkspaceStubModel) Init() tea.Cmd {
	return nil
}

// Update handles keypress events on the dashboard.
func (m WorkspaceStubModel) Update(msg tea.Msg) (WorkspaceStubModel, tea.Cmd) {
	return m, nil
}

// View renders the Phase 1 completion dashboard.
func (m WorkspaceStubModel) View() string {
	// 1. Header
	logo := styles.LogoStyle.Render(styles.AsciiLogo)
	subtitle := styles.SubtitleStyle.Render("Terminal Realtime Workspace")
	header := lipgloss.JoinVertical(lipgloss.Center, logo, subtitle)

	// 2. User info box
	name := m.cfg.User.Name
	email := m.cfg.User.Email
	id := m.cfg.User.ID

	if m.user != nil {
		if m.user.Name != "" {
			name = m.user.Name
		}
		if m.user.Email != "" {
			email = m.user.Email
		}
		if m.user.ID != "" {
			id = m.user.ID
		}
	}

	userInfoTitle := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorAccent).Render("Authenticated Session")
	nameLine := fmt.Sprintf("%s %s", styles.LabelStyle.Render("Name: "), name)
	emailLine := fmt.Sprintf("%s %s", styles.LabelStyle.Render("Email:"), email)
	idLine := fmt.Sprintf("%s %s", styles.LabelStyle.Render("ID:   "), styles.HelpStyle.Render(id))

	userInfoBox := lipgloss.JoinVertical(lipgloss.Left, userInfoTitle, "\n", nameLine, emailLine, idLine)

	// 3. Phase 1 Success Banner & Next Steps
	statusBox := styles.SuccessBoxStyle.Render("✔ Phase 1 Complete: Authentication & Session Management Active")

	phase2Notice := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorSecondary).
		Padding(1, 2).
		Width(60).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSecondary).Render("Next Step: Phase 2 (Workspace Management)"),
				"\n",
				styles.HelpStyle.Render("• Browse and search public workspaces"),
				styles.HelpStyle.Render("• Create new workspace tenants with subdomain routing"),
				styles.HelpStyle.Render("• Join workspaces and manage invitations"),
			),
		)

	// 4. Keyboard Shortcuts
	footer := styles.HelpStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			styles.HelpKeyStyle.Render("Ctrl+L: "), "Log Out & Clear Session   ",
			styles.HelpKeyStyle.Render("Ctrl+C / q: "), "Quit",
		),
	)

	// 5. Assemble and Center
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		"\n",
		styles.CardStyle.Render(userInfoBox),
		"\n",
		statusBox,
		"\n",
		phase2Notice,
		"\n",
		footer,
	)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}

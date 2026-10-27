package views

import (
	"context"
	"fmt"
	"strings"

	"Conclave-TUI/internal/api"
	"Conclave-TUI/internal/config"
	"Conclave-TUI/internal/tui/common"
	"Conclave-TUI/internal/tui/styles"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// WorkspaceTab defines the active section in the workspace manager.
type WorkspaceTab int

const (
	TabMyWorkspaces WorkspaceTab = iota
	TabSearch
	TabCreate
	TabInvitations
)

// WorkspaceModel coordinates workspace discovery, creation, and selection.
type WorkspaceModel struct {
	cfg    *config.Config
	client *api.Client

	// Active tab
	tab WorkspaceTab

	// State data
	myWorkspaces []api.Tenant
	myCursor     int

	searchResults []api.Tenant
	searchCursor  int
	searchInput   textinput.Model

	createNameInput textinput.Model
	createIsPublic  bool
	createFocus     int // 0: Name, 1: Checkbox, 2: Button

	invitations     []api.TenantInvitation
	inviteCursor    int

	// Active workspace view (once selected)
	activeTenant *api.Tenant
	tenantDetail *api.TenantDetail
	members      []api.Member

	// Status & loading
	loading bool
	message string
	isError bool
	spinner spinner.Model

	// Viewport dimensions
	width  int
	height int
}

// NewWorkspaceModel initializes the workspace manager screen.
func NewWorkspaceModel(cfg *config.Config, client *api.Client) WorkspaceModel {
	// 1. Search text input
	si := textinput.New()
	si.Placeholder = "Search public workspaces by name..."
	si.CharLimit = 50
	si.SetWidth(40)

	// 2. Create workspace inputs
	ci := textinput.New()
	ci.Placeholder = "workspace-name (e.g. engineering)"
	ci.CharLimit = 50
	ci.SetWidth(40)

	// 3. Spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorSecondary)

	return WorkspaceModel{
		cfg:             cfg,
		client:          client,
		tab:             TabMyWorkspaces,
		searchInput:     si,
		createNameInput: ci,
		createIsPublic:  true,
		spinner:         sp,
	}
}

// SetSize updates layout dimensions.
func (m *WorkspaceModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Init loads the user's workspaces and pending invitations on startup.
func (m WorkspaceModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadMyWorkspacesCmd(),
		m.loadInvitationsCmd(),
		textinput.Blink,
	)
}

// ActiveTenant returns the currently entered tenant (if any).
func (m WorkspaceModel) ActiveTenant() *api.Tenant {
	return m.activeTenant
}

// SetActiveTenant sets the selected tenant directly (e.g. from config).
func (m *WorkspaceModel) SetActiveTenant(t *api.Tenant) tea.Cmd {
	m.activeTenant = t
	if t != nil {
		return m.loadTenantDetailsCmd(t.Name)
	}
	return nil
}

// ClearActiveTenant returns to the workspace selection list.
func (m *WorkspaceModel) ClearActiveTenant() {
	m.activeTenant = nil
	m.tenantDetail = nil
	m.members = nil
}

// loadMyWorkspacesCmd fetches GET /tenant/mine/.
func (m WorkspaceModel) loadMyWorkspacesCmd() tea.Cmd {
	return func() tea.Msg {
		tenants, err := m.client.GetMyTenants(context.Background())
		if err != nil {
			return common.WorkspaceErrorMsg{Err: err}
		}
		return common.WorkspaceListLoadedMsg{Tenants: tenants}
	}
}

// searchTenantsCmd calls GET /tenant/search/?q=.
func (m WorkspaceModel) searchTenantsCmd(query string) tea.Cmd {
	return func() tea.Msg {
		results, err := m.client.SearchTenants(context.Background(), query)
		if err != nil {
			return common.WorkspaceErrorMsg{Err: err}
		}
		return common.SearchResultsLoadedMsg{Query: query, Results: results}
	}
}

// createTenantCmd calls POST /tenant/.
func (m WorkspaceModel) createTenantCmd(name string, isPublic bool) tea.Cmd {
	return func() tea.Msg {
		tenant, err := m.client.CreateTenant(context.Background(), name, isPublic)
		if err != nil {
			return common.WorkspaceErrorMsg{Err: err}
		}
		return common.WorkspaceCreatedMsg{Tenant: tenant}
	}
}

// joinTenantCmd calls POST /tenant/join/.
func (m WorkspaceModel) joinTenantCmd(name string) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.client.JoinTenant(context.Background(), name)
		if err != nil {
			return common.WorkspaceErrorMsg{Err: err}
		}
		return common.WorkspaceJoinedMsg{Response: resp, TenantName: name}
	}
}

// loadInvitationsCmd calls GET /tenant/invitations/.
func (m WorkspaceModel) loadInvitationsCmd() tea.Cmd {
	return func() tea.Msg {
		invites, err := m.client.GetMyInvitations(context.Background())
		if err != nil {
			return common.WorkspaceErrorMsg{Err: err}
		}
		return common.InvitationsLoadedMsg{Invitations: invites}
	}
}

// respondInvitationCmd calls POST /tenant/invitations/<id>/respond/.
func (m WorkspaceModel) respondInvitationCmd(id string, accept bool) tea.Cmd {
	return func() tea.Msg {
		err := m.client.RespondInvitation(context.Background(), id, accept)
		if err != nil {
			return common.WorkspaceErrorMsg{Err: err}
		}
		action := "accepted"
		if !accept {
			action = "declined"
		}
		return common.InvitationRespondedMsg{InvitationID: id, Action: action}
	}
}

// loadTenantDetailsCmd fetches subdomain overview and member list.
func (m WorkspaceModel) loadTenantDetailsCmd(tenantName string) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			detail, err := m.client.GetTenantDetail(context.Background(), tenantName)
			if err != nil {
				return common.WorkspaceErrorMsg{Err: err}
			}
			return common.WorkspaceDetailLoadedMsg{Detail: detail}
		},
		func() tea.Msg {
			members, err := m.client.GetTenantMembers(context.Background(), tenantName)
			if err != nil {
				return common.WorkspaceErrorMsg{Err: err}
			}
			return members
		},
	)
}

// Update coordinates incoming keystrokes and async messages.
func (m WorkspaceModel) Update(msg tea.Msg) (WorkspaceModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// 1. Data loaded messages
	case common.WorkspaceListLoadedMsg:
		m.loading = false
		m.myWorkspaces = msg.Tenants
		if m.myCursor >= len(m.myWorkspaces) {
			m.myCursor = max(0, len(m.myWorkspaces)-1)
		}

	case common.SearchResultsLoadedMsg:
		m.loading = false
		m.searchResults = msg.Results
		m.searchCursor = 0

	case common.WorkspaceCreatedMsg:
		m.loading = false
		m.message = fmt.Sprintf("Successfully created workspace %q!", msg.Tenant.Name)
		m.isError = false
		m.myWorkspaces = append(m.myWorkspaces, *msg.Tenant)
		m.tab = TabMyWorkspaces
		m.createNameInput.Reset()
		cmds = append(cmds, m.loadMyWorkspacesCmd())

	case common.WorkspaceJoinedMsg:
		m.loading = false
		m.message = msg.Response.Message
		m.isError = false
		cmds = append(cmds, m.loadMyWorkspacesCmd())

	case common.InvitationsLoadedMsg:
		m.invitations = msg.Invitations
		if m.inviteCursor >= len(m.invitations) {
			m.inviteCursor = max(0, len(m.invitations)-1)
		}

	case common.InvitationRespondedMsg:
		m.loading = false
		m.message = fmt.Sprintf("Invitation %s.", msg.Action)
		m.isError = false
		cmds = append(cmds, m.loadInvitationsCmd(), m.loadMyWorkspacesCmd())

	case common.WorkspaceDetailLoadedMsg:
		m.loading = false
		m.tenantDetail = msg.Detail

	case []api.Member:
		m.members = msg

	case common.WorkspaceErrorMsg:
		m.loading = false
		m.message = msg.Err.Error()
		m.isError = true

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	// 2. Keyboard interaction
	case tea.KeyPressMsg:
		key := msg.String()

		// If viewing an active workspace dashboard:
		if m.activeTenant != nil {
			switch key {
			case "esc", "w", "backspace":
				m.ClearActiveTenant()
				m.message = ""
				return m, m.loadMyWorkspacesCmd()
			}
			return m, nil
		}

		// Tab navigation via 1, 2, 3, 4
		switch key {
		case "1":
			m.tab = TabMyWorkspaces
			m.message = ""
			return m, m.loadMyWorkspacesCmd()
		case "2":
			m.tab = TabSearch
			m.message = ""
			m.searchInput.Focus()
			return m, m.searchTenantsCmd(m.searchInput.Value())
		case "3":
			m.tab = TabCreate
			m.message = ""
			m.createNameInput.Focus()
			m.createFocus = 0
			return m, nil
		case "4":
			m.tab = TabInvitations
			m.message = ""
			return m, m.loadInvitationsCmd()
		}

		// Handle keypresses per tab
		switch m.tab {
		case TabMyWorkspaces:
			switch key {
			case "up", "k":
				if m.myCursor > 0 {
					m.myCursor--
				}
			case "down", "j":
				if m.myCursor < len(m.myWorkspaces)-1 {
					m.myCursor++
				}
			case "enter":
				if len(m.myWorkspaces) > 0 {
					selected := m.myWorkspaces[m.myCursor]
					m.activeTenant = &selected
					m.cfg.ActiveTenant = selected.Name
					_ = m.cfg.Save()
					cmds = append(cmds, m.loadTenantDetailsCmd(selected.Name))
					cmds = append(cmds, func() tea.Msg {
						return common.WorkspaceSelectedMsg{Tenant: selected}
					})
				}
			case "r":
				m.loading = true
				cmds = append(cmds, m.spinner.Tick, m.loadMyWorkspacesCmd())
			}

		case TabSearch:
			switch key {
			case "up":
				if m.searchCursor > 0 {
					m.searchCursor--
				}
			case "down":
				if m.searchCursor < len(m.searchResults)-1 {
					m.searchCursor++
				}
			case "enter":
				// If typing in search box, execute search
				if m.searchInput.Focused() && len(m.searchResults) == 0 {
					m.loading = true
					cmds = append(cmds, m.spinner.Tick, m.searchTenantsCmd(m.searchInput.Value()))
				} else if len(m.searchResults) > 0 {
					// Join selected workspace
					target := m.searchResults[m.searchCursor]
					m.loading = true
					cmds = append(cmds, m.spinner.Tick, m.joinTenantCmd(target.Name))
				}
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				cmds = append(cmds, cmd)
			}

		case TabCreate:
			switch key {
			case "tab", "down":
				m.createFocus = (m.createFocus + 1) % 3
				m.updateCreateFocus()
			case "shift+tab", "up":
				m.createFocus--
				if m.createFocus < 0 {
					m.createFocus = 2
				}
				m.updateCreateFocus()
			case " ":
				if m.createFocus == 1 {
					m.createIsPublic = !m.createIsPublic
				}
			case "enter":
				if m.createFocus == 2 || m.createFocus == 0 {
					name := strings.TrimSpace(m.createNameInput.Value())
					if name != "" {
						m.loading = true
						cmds = append(cmds, m.spinner.Tick, m.createTenantCmd(name, m.createIsPublic))
					}
				}
			default:
				if m.createFocus == 0 {
					var cmd tea.Cmd
					m.createNameInput, cmd = m.createNameInput.Update(msg)
					cmds = append(cmds, cmd)
				}
			}

		case TabInvitations:
			switch key {
			case "up", "k":
				if m.inviteCursor > 0 {
					m.inviteCursor--
				}
			case "down", "j":
				if m.inviteCursor < len(m.invitations)-1 {
					m.inviteCursor++
				}
			case "a":
				if len(m.invitations) > 0 {
					inv := m.invitations[m.inviteCursor]
					m.loading = true
					cmds = append(cmds, m.spinner.Tick, m.respondInvitationCmd(inv.ID, true))
				}
			case "d":
				if len(m.invitations) > 0 {
					inv := m.invitations[m.inviteCursor]
					m.loading = true
					cmds = append(cmds, m.spinner.Tick, m.respondInvitationCmd(inv.ID, false))
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *WorkspaceModel) updateCreateFocus() {
	if m.createFocus == 0 {
		m.createNameInput.Focus()
	} else {
		m.createNameInput.Blur()
	}
}

// View renders the workspace manager interface.
func (m WorkspaceModel) View() string {
	// If currently viewing an entered active workspace:
	if m.activeTenant != nil {
		return m.renderActiveWorkspaceView()
	}

	// 1. Header & Top Status
	logo := styles.LogoStyle.Render(styles.AsciiLogo)
	userLine := fmt.Sprintf("Logged in as: %s (%s)", m.cfg.User.Name, m.cfg.User.Email)
	header := lipgloss.JoinVertical(
		lipgloss.Center,
		logo,
		styles.SubtitleStyle.Render(userLine),
	)

	// 2. Navigation Tabs
	var tabs []string
	myTabLabel := fmt.Sprintf("1. My Workspaces (%d)", len(m.myWorkspaces))
	searchTabLabel := "2. Explore & Join"
	createTabLabel := "3. Create Workspace"
	invTabLabel := fmt.Sprintf("4. Invitations (%d)", len(m.invitations))

	if m.tab == TabMyWorkspaces {
		tabs = append(tabs, styles.ActiveTabStyle.Render(myTabLabel))
	} else {
		tabs = append(tabs, styles.InactiveTabStyle.Render(myTabLabel))
	}

	if m.tab == TabSearch {
		tabs = append(tabs, styles.ActiveTabStyle.Render(searchTabLabel))
	} else {
		tabs = append(tabs, styles.InactiveTabStyle.Render(searchTabLabel))
	}

	if m.tab == TabCreate {
		tabs = append(tabs, styles.ActiveTabStyle.Render(createTabLabel))
	} else {
		tabs = append(tabs, styles.InactiveTabStyle.Render(createTabLabel))
	}

	if m.tab == TabInvitations {
		tabs = append(tabs, styles.ActiveTabStyle.Render(invTabLabel))
	} else {
		tabs = append(tabs, styles.InactiveTabStyle.Render(invTabLabel))
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top,
		tabs[0], styles.TabDividerStyle.Render(" "),
		tabs[1], styles.TabDividerStyle.Render(" "),
		tabs[2], styles.TabDividerStyle.Render(" "),
		tabs[3],
	)

	// 3. Status/Notification Banner
	var statusBanner string
	if m.message != "" {
		if m.isError {
			statusBanner = styles.ErrorBoxStyle.Render("✕ " + m.message)
		} else {
			statusBanner = styles.SuccessBoxStyle.Render("✔ " + m.message)
		}
	} else if m.loading {
		statusBanner = lipgloss.JoinHorizontal(
			lipgloss.Center,
			m.spinner.View(),
			" ",
			styles.SubtitleStyle.Render("Connecting to Conclave..."),
		)
	}

	// 4. Tab Body Content
	var tabBody string
	switch m.tab {
	case TabMyWorkspaces:
		tabBody = m.renderMyWorkspacesTab()
	case TabSearch:
		tabBody = m.renderSearchTab()
	case TabCreate:
		tabBody = m.renderCreateTab()
	case TabInvitations:
		tabBody = m.renderInvitationsTab()
	}

	// 5. Help Footer
	footer := m.renderFooter()

	// 6. Assemble Layout
	contentElements := []string{
		header,
		"\n",
		tabBar,
	}
	if statusBanner != "" {
		contentElements = append(contentElements, statusBanner)
	}
	contentElements = append(contentElements, "\n", tabBody, "\n", footer)

	content := lipgloss.JoinVertical(lipgloss.Center, contentElements...)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}

func (m WorkspaceModel) renderMyWorkspacesTab() string {
	if len(m.myWorkspaces) == 0 {
		emptyMsg := lipgloss.JoinVertical(
			lipgloss.Center,
			styles.SubtitleStyle.Render("You haven't joined any workspaces yet."),
			"\n",
			styles.HelpStyle.Render("• Press [2] to discover and join public workspaces"),
			styles.HelpStyle.Render("• Press [3] to create a new workspace for your team"),
		)
		return styles.CardStyle.Render(emptyMsg)
	}

	var items []string
	for i, ws := range m.myWorkspaces {
		privacyBadge := styles.BadgePublicStyle.Render("[Public]")
		if !ws.IsPublic {
			privacyBadge = styles.BadgePrivateStyle.Render("[Private]")
		}

		memberText := styles.MemberCountStyle.Render(fmt.Sprintf("%d members", ws.MemberCount))

		prefix := "  "
		if i == m.myCursor {
			prefix = "▸ "
		}

		nameStyled := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorForeground).Render(ws.Name)
		if i == m.myCursor {
			nameStyled = lipgloss.NewStyle().Bold(true).Foreground(styles.ColorAccent).Render(ws.Name)
		}

		line := fmt.Sprintf("%s%s   %s   %s", prefix, nameStyled, privacyBadge, memberText)

		if i == m.myCursor {
			items = append(items, styles.SelectedItemStyle.Width(58).Render(line))
		} else {
			items = append(items, styles.NormalItemStyle.Width(58).Render(line))
		}
	}

	listContainer := lipgloss.JoinVertical(lipgloss.Left, items...)
	return styles.CardStyle.Render(listContainer)
}

func (m WorkspaceModel) renderSearchTab() string {
	searchBox := styles.InputContainerStyle.Render(m.searchInput.View())

	var results []string
	if len(m.searchResults) == 0 {
		results = append(results, styles.HelpStyle.Render("Type a name and press Enter to search public workspaces."))
	} else {
		for i, ws := range m.searchResults {
			prefix := "  "
			if i == m.searchCursor {
				prefix = "▸ "
			}

			// Check if already in my workspaces
			alreadyJoined := false
			for _, my := range m.myWorkspaces {
				if my.ID == ws.ID {
					alreadyJoined = true
					break
				}
			}

			badge := styles.BadgePublicStyle.Render("[Public]")
			if alreadyJoined {
				badge = styles.BadgeJoinedStyle.Render("[Joined]")
			}

			memberText := styles.MemberCountStyle.Render(fmt.Sprintf("%d members", ws.MemberCount))
			line := fmt.Sprintf("%s%s   %s   %s", prefix, ws.Name, badge, memberText)

			if i == m.searchCursor {
				results = append(results, styles.SelectedItemStyle.Width(58).Render(line))
			} else {
				results = append(results, styles.NormalItemStyle.Width(58).Render(line))
			}
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.LabelStyle.Render("Search Directory:"),
		searchBox,
		"\n",
		lipgloss.JoinVertical(lipgloss.Left, results...),
	)
	return styles.CardStyle.Render(content)
}

func (m WorkspaceModel) renderCreateTab() string {
	nameLabel := styles.LabelStyle.Render("Workspace Name (subdomain identifier):")
	if m.createFocus == 0 {
		nameLabel = styles.FocusedLabelStyle.Render("Workspace Name (subdomain identifier):")
	}

	nameBox := styles.InputContainerStyle.Render(m.createNameInput.View())
	if m.createFocus == 0 {
		nameBox = styles.FocusedInputContainerStyle.Render(m.createNameInput.View())
	}

	checkbox := "[ ] Private Workspace (Invite only)"
	if m.createIsPublic {
		checkbox = "[x] Public Workspace (Discoverable by everyone)"
	}
	checkboxStyle := styles.LabelStyle
	if m.createFocus == 1 {
		checkboxStyle = styles.FocusedLabelStyle
	}
	checkboxLine := checkboxStyle.Render(checkbox) + " " + styles.HelpStyle.Render("(Press Space to toggle)")

	btnText := "Create Workspace →"
	var btn string
	if m.createFocus == 2 {
		btn = styles.ActiveButtonStyle.Render(btnText)
	} else {
		btn = styles.InactiveButtonStyle.Render(btnText)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		nameLabel,
		nameBox,
		"\n",
		checkboxLine,
		"\n",
		btn,
	)
	return styles.CardStyle.Render(content)
}

func (m WorkspaceModel) renderInvitationsTab() string {
	if len(m.invitations) == 0 {
		return styles.CardStyle.Render(styles.HelpStyle.Render("No pending workspace invitations."))
	}

	var items []string
	for i, inv := range m.invitations {
		prefix := "  "
		if i == m.inviteCursor {
			prefix = "▸ "
		}

		line := fmt.Sprintf("%s%s (from %s)", prefix, inv.TenantName, inv.InvitedByName)
		hint := styles.HelpStyle.Render("[a] Accept   [d] Decline")

		block := lipgloss.JoinVertical(lipgloss.Left, line, hint)
		if i == m.inviteCursor {
			items = append(items, styles.SelectedItemStyle.Width(58).Render(block))
		} else {
			items = append(items, styles.NormalItemStyle.Width(58).Render(block))
		}
	}

	return styles.CardStyle.Render(lipgloss.JoinVertical(lipgloss.Left, items...))
}

func (m WorkspaceModel) renderActiveWorkspaceView() string {
	ws := m.activeTenant
	if ws == nil {
		return ""
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorAccent).Render(fmt.Sprintf("⬢ Workspace: %s", ws.Name))
	subdomain := fmt.Sprintf("%s.lvh.me:8000", ws.Name)
	subdomainLine := styles.SubtitleStyle.Render(fmt.Sprintf("Routing Host: %s", subdomain))

	var infoLines []string
	infoLines = append(infoLines, title, subdomainLine, "\n")

	if m.tenantDetail != nil {
		infoLines = append(infoLines, fmt.Sprintf("%s %s", styles.LabelStyle.Render("Owner:       "), m.tenantDetail.Owner))
		infoLines = append(infoLines, fmt.Sprintf("%s %d", styles.LabelStyle.Render("Total Members:"), m.tenantDetail.MemberCount))
		infoLines = append(infoLines, fmt.Sprintf("%s %d", styles.LabelStyle.Render("Chat Channels:"), m.tenantDetail.RoomCount))
	} else {
		infoLines = append(infoLines, fmt.Sprintf("%s %d", styles.LabelStyle.Render("Total Members:"), ws.MemberCount))
	}

	// Live member roster
	if len(m.members) > 0 {
		infoLines = append(infoLines, "\n", styles.FocusedLabelStyle.Render("Active Member Roster:"))
		for _, member := range m.members {
			statusDot := styles.HelpStyle.Render("○ offline")
			if member.IsOnline {
				statusDot = styles.BadgePublicStyle.Render("● online")
			}
			roleTag := fmt.Sprintf("[%s]", member.Role)
			memberLine := fmt.Sprintf("  • %-16s %-10s %s", member.Name, roleTag, statusDot)
			infoLines = append(infoLines, memberLine)
		}
	}

	cardContent := lipgloss.JoinVertical(lipgloss.Left, infoLines...)

	readyNotice := styles.SuccessBoxStyle.Render("✔ Workspace Selected: Ready for Phase 3 (Chat Core & WebSocket Streaming)")

	footer := styles.HelpStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			styles.HelpKeyStyle.Render("Esc / Back: "), "Switch Workspace   ",
			styles.HelpKeyStyle.Render("Ctrl+L: "), "Log Out   ",
			styles.HelpKeyStyle.Render("Ctrl+C / q: "), "Quit",
		),
	)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.LogoStyle.Render(styles.AsciiLogo),
		"\n",
		styles.CardStyle.Render(cardContent),
		"\n",
		readyNotice,
		"\n",
		footer,
	)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}

func (m WorkspaceModel) renderFooter() string {
	switch m.tab {
	case TabMyWorkspaces:
		return styles.HelpStyle.Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				styles.HelpKeyStyle.Render("1-4: "), "Switch Tab   ",
				styles.HelpKeyStyle.Render("↑/↓: "), "Navigate   ",
				styles.HelpKeyStyle.Render("Enter: "), "Enter Workspace   ",
				styles.HelpKeyStyle.Render("r: "), "Refresh   ",
				styles.HelpKeyStyle.Render("Ctrl+L: "), "Log Out",
			),
		)
	case TabSearch:
		return styles.HelpStyle.Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				styles.HelpKeyStyle.Render("1-4: "), "Switch Tab   ",
				styles.HelpKeyStyle.Render("Enter: "), "Search / Join Workspace   ",
				styles.HelpKeyStyle.Render("Ctrl+L: "), "Log Out",
			),
		)
	case TabCreate:
		return styles.HelpStyle.Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				styles.HelpKeyStyle.Render("1-4: "), "Switch Tab   ",
				styles.HelpKeyStyle.Render("Tab/↓: "), "Next Field   ",
				styles.HelpKeyStyle.Render("Space: "), "Toggle Public/Private   ",
				styles.HelpKeyStyle.Render("Enter: "), "Create",
			),
		)
	case TabInvitations:
		return styles.HelpStyle.Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				styles.HelpKeyStyle.Render("1-4: "), "Switch Tab   ",
				styles.HelpKeyStyle.Render("a: "), "Accept Invitation   ",
				styles.HelpKeyStyle.Render("d: "), "Decline Invitation",
			),
		)
	}
	return ""
}

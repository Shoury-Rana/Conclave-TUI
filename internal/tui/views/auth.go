package views

import (
	"context"
	"strings"

	"Conclave-TUI/internal/api"
	"Conclave-TUI/internal/tui/common"
	"Conclave-TUI/internal/tui/styles"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// AuthMode represents whether the auth view is in Login or Sign Up mode.
type AuthMode int

const (
	ModeLogin AuthMode = iota
	ModeSignup
)

// AuthModel manages the state and rendering for the authentication screen.
type AuthModel struct {
	client *api.Client

	mode         AuthMode
	focusIndex   int
	loading      bool
	errorMessage string

	// Form input components
	nameInput     textinput.Model
	emailInput    textinput.Model
	passwordInput textinput.Model

	// Spinner for async network activity
	spinner spinner.Model

	// Window dimensions for centering
	width  int
	height int
}

// NewAuthModel initializes the authentication model with styled text inputs.
func NewAuthModel(client *api.Client) AuthModel {
	// 1. Name input (for Sign Up)
	name := textinput.New()
	name.Placeholder = "Alex Mercer"
	name.CharLimit = 150
	name.SetWidth(40)

	// 2. Email input
	email := textinput.New()
	email.Placeholder = "alex@conclave.local"
	email.CharLimit = 150
	email.SetWidth(40)
	email.Focus() // Start with email focused in Login mode

	// 3. Password input
	pwd := textinput.New()
	pwd.Placeholder = "••••••••"
	pwd.EchoMode = textinput.EchoPassword
	pwd.EchoCharacter = '•'
	pwd.CharLimit = 100
	pwd.SetWidth(40)

	// 4. Loading Spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorSecondary)

	return AuthModel{
		client:        client,
		mode:          ModeLogin,
		focusIndex:    0, // Points to email in Login mode
		nameInput:     name,
		emailInput:    email,
		passwordInput: pwd,
		spinner:       sp,
	}
}

// Init implements tea.Model and starts the input blinking cursor.
func (m AuthModel) Init() tea.Cmd {
	return textinput.Blink
}

// SetSize updates the viewport dimensions for responsive centering.
func (m *AuthModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// SetLoading toggles the loading state and error message.
func (m *AuthModel) SetLoading(loading bool) {
	m.loading = loading
	if loading {
		m.errorMessage = ""
	}
}

// SetError updates the displayed error banner.
func (m *AuthModel) SetError(err error) {
	m.loading = false
	if err != nil {
		m.errorMessage = err.Error()
	} else {
		m.errorMessage = ""
	}
}

// Update handles incoming user keystrokes and async events.
func (m AuthModel) Update(msg tea.Msg) (AuthModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		key := msg.String()

		// Toggle Login/Signup modes via Ctrl+T
		if key == "ctrl+t" {
			m.toggleMode()
			return m, nil
		}

		// If currently loading, block form navigation
		if m.loading {
			break
		}

		switch key {
		case "tab", "down":
			m.nextInput()
			return m, nil

		case "shift+tab", "up":
			m.prevInput()
			return m, nil

		case "enter":
			// If on the submit button or last input field, submit form
			if m.isSubmitFocused() {
				return m, m.submitForm()
			}
			// Otherwise move to the next field
			m.nextInput()
			return m, nil
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Update text inputs if not loading
	if !m.loading {
		var cmd tea.Cmd
		if m.mode == ModeSignup {
			m.nameInput, cmd = m.nameInput.Update(msg)
			cmds = append(cmds, cmd)
		}
		m.emailInput, cmd = m.emailInput.Update(msg)
		cmds = append(cmds, cmd)

		m.passwordInput, cmd = m.passwordInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// toggleMode switches between Login and Sign Up.
func (m *AuthModel) toggleMode() {
	if m.mode == ModeLogin {
		m.mode = ModeSignup
		m.focusIndex = 0
		m.nameInput.Focus()
		m.emailInput.Blur()
		m.passwordInput.Blur()
	} else {
		m.mode = ModeLogin
		m.focusIndex = 0
		m.nameInput.Blur()
		m.emailInput.Focus()
		m.passwordInput.Blur()
	}
	m.errorMessage = ""
}

// maxIndex returns the index of the submit button for the active mode.
func (m AuthModel) maxIndex() int {
	if m.mode == ModeLogin {
		return 2 // 0: Email, 1: Password, 2: Submit
	}
	return 3 // 0: Name, 1: Email, 2: Password, 3: Submit
}

func (m AuthModel) isSubmitFocused() bool {
	return m.focusIndex == m.maxIndex()
}

func (m *AuthModel) nextInput() {
	m.focusIndex = (m.focusIndex + 1) % (m.maxIndex() + 1)
	m.updateFocus()
}

func (m *AuthModel) prevInput() {
	m.focusIndex--
	if m.focusIndex < 0 {
		m.focusIndex = m.maxIndex()
	}
	m.updateFocus()
}

func (m *AuthModel) updateFocus() {
	if m.mode == ModeLogin {
		if m.focusIndex == 0 {
			m.emailInput.Focus()
			m.passwordInput.Blur()
		} else if m.focusIndex == 1 {
			m.emailInput.Blur()
			m.passwordInput.Focus()
		} else {
			m.emailInput.Blur()
			m.passwordInput.Blur()
		}
	} else {
		if m.focusIndex == 0 {
			m.nameInput.Focus()
			m.emailInput.Blur()
			m.passwordInput.Blur()
		} else if m.focusIndex == 1 {
			m.nameInput.Blur()
			m.emailInput.Focus()
			m.passwordInput.Blur()
		} else if m.focusIndex == 2 {
			m.nameInput.Blur()
			m.emailInput.Blur()
			m.passwordInput.Focus()
		} else {
			m.nameInput.Blur()
			m.emailInput.Blur()
			m.passwordInput.Blur()
		}
	}
}

// submitForm initiates the async Login or Signup API call as a tea.Cmd.
func (m *AuthModel) submitForm() tea.Cmd {
	m.loading = true
	m.errorMessage = ""

	email := strings.TrimSpace(m.emailInput.Value())
	password := m.passwordInput.Value()

	if m.mode == ModeLogin {
		return tea.Batch(
			m.spinner.Tick,
			func() tea.Msg {
				resp, err := m.client.Login(context.Background(), email, password)
				if err != nil {
					return common.AuthErrorMsg{Err: err}
				}
				return common.LoginSuccessMsg{Response: resp}
			},
		)
	}

	name := strings.TrimSpace(m.nameInput.Value())
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			resp, err := m.client.Signup(context.Background(), name, email, password)
			if err != nil {
				return common.AuthErrorMsg{Err: err}
			}
			return common.SignupSuccessMsg{Response: resp}
		},
	)
}

// View renders the complete authentication interface.
func (m AuthModel) View() string {
	// 1. Render Logo and Header
	logo := styles.LogoStyle.Render(styles.AsciiLogo)
	subtitle := styles.SubtitleStyle.Render("Terminal Workspace & Chat Platform")
	header := lipgloss.JoinVertical(lipgloss.Center, logo, subtitle)

	// 2. Render Tabs (Login / Sign Up)
	var loginTab, signupTab string
	if m.mode == ModeLogin {
		loginTab = styles.ActiveTabStyle.Render("1. Login")
		signupTab = styles.InactiveTabStyle.Render("2. Sign Up")
	} else {
		loginTab = styles.InactiveTabStyle.Render("1. Login")
		signupTab = styles.ActiveTabStyle.Render("2. Sign Up")
	}
	tabs := lipgloss.JoinHorizontal(lipgloss.Top, loginTab, styles.TabDividerStyle.Render(" "), signupTab)

	// 3. Render Form Fields
	var formFields []string

	if m.mode == ModeSignup {
		nameLabel := styles.LabelStyle.Render("Full Name")
		if m.focusIndex == 0 {
			nameLabel = styles.FocusedLabelStyle.Render("Full Name")
		}
		nameBox := styles.InputContainerStyle.Render(m.nameInput.View())
		if m.focusIndex == 0 {
			nameBox = styles.FocusedInputContainerStyle.Render(m.nameInput.View())
		}
		formFields = append(formFields, nameLabel, nameBox)
	}

	emailIdx := 0
	if m.mode == ModeSignup {
		emailIdx = 1
	}
	emailLabel := styles.LabelStyle.Render("Email Address")
	if m.focusIndex == emailIdx {
		emailLabel = styles.FocusedLabelStyle.Render("Email Address")
	}
	emailBox := styles.InputContainerStyle.Render(m.emailInput.View())
	if m.focusIndex == emailIdx {
		emailBox = styles.FocusedInputContainerStyle.Render(m.emailInput.View())
	}
	formFields = append(formFields, emailLabel, emailBox)

	pwdIdx := 1
	if m.mode == ModeSignup {
		pwdIdx = 2
	}
	pwdLabel := styles.LabelStyle.Render("Password")
	if m.focusIndex == pwdIdx {
		pwdLabel = styles.FocusedLabelStyle.Render("Password")
	}
	pwdBox := styles.InputContainerStyle.Render(m.passwordInput.View())
	if m.focusIndex == pwdIdx {
		pwdBox = styles.FocusedInputContainerStyle.Render(m.passwordInput.View())
	}
	formFields = append(formFields, pwdLabel, pwdBox)

	// 4. Render Submit Button or Spinner
	var submitButton string
	if m.loading {
		submitButton = lipgloss.JoinHorizontal(
			lipgloss.Center,
			m.spinner.View(),
			" ",
			styles.SubtitleStyle.Render("Authenticating with Conclave..."),
		)
	} else {
		btnText := "Log In →"
		if m.mode == ModeSignup {
			btnText = "Create Account →"
		}
		if m.isSubmitFocused() {
			submitButton = styles.ActiveButtonStyle.Render(btnText)
		} else {
			submitButton = styles.InactiveButtonStyle.Render(btnText)
		}
	}
	formFields = append(formFields, submitButton)

	// 5. Render Error Message (if any)
	if m.errorMessage != "" {
		errorView := styles.ErrorBoxStyle.Render("✕ " + m.errorMessage)
		formFields = append([]string{errorView}, formFields...)
	}

	formContent := lipgloss.JoinVertical(lipgloss.Left, formFields...)

	// 6. Wrap in Card
	card := styles.CardStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			tabs,
			"\n",
			formContent,
		),
	)

	// 7. Footer / Keyboard Help
	help := styles.HelpStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			styles.HelpKeyStyle.Render("Tab/↓: "), "Next Field   ",
			styles.HelpKeyStyle.Render("Enter: "), "Submit   ",
			styles.HelpKeyStyle.Render("Ctrl+T: "), "Toggle Mode   ",
			styles.HelpKeyStyle.Render("Ctrl+C: "), "Quit",
		),
	)

	// 8. Center everything in the terminal
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		"\n",
		card,
		"\n",
		help,
	)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}

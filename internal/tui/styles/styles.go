package styles

import (
	"charm.land/lipgloss/v2"
)

// Color Palette — Modern, sleek dark theme
var (
	// Brand & Accents
	ColorPrimary   = lipgloss.Color("#7C3AED") // Deep Violet / Indigo
	ColorSecondary = lipgloss.Color("#06B6D4") // Vibrant Cyan
	ColorAccent    = lipgloss.Color("#A78BFA") // Light Lavender Accent

	// Text & Backgrounds
	ColorForeground = lipgloss.Color("#F9FAFB") // Crisp White/Off-white
	ColorMuted      = lipgloss.Color("#9CA3AF") // Cool Gray
	ColorDimmed     = lipgloss.Color("#6B7280") // Dimmed Gray
	ColorDarkBg     = lipgloss.Color("#0F172A") // Slate Dark Background

	// Status Colors
	ColorSuccess = lipgloss.Color("#10B981") // Emerald Green
	ColorError   = lipgloss.Color("#F43F5E") // Rose Red
	ColorWarning = lipgloss.Color("#F59E0B") // Amber Yellow

	// Border Colors
	ColorBorderDefault = lipgloss.Color("#374151") // Slate Border
	ColorBorderFocused = lipgloss.Color("#8B5CF6") // Highlighted Violet Border
)

// UI Component Styles
var (
	// Logo & Titles
	LogoStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Italic(true)

	// App Container / Shell
	AppStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Card Box (Used for Auth form and Dialogs)
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorderFocused).
			Padding(1, 3).
			Width(64)

	// Tabs for switching between Login and Signup
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorForeground).
			Background(ColorPrimary).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Background(lipgloss.Color("#1F2937")).
				Padding(0, 2)

	TabDividerStyle = lipgloss.NewStyle().
			Foreground(ColorDimmed).
			Padding(0, 1)

	// Form Elements
	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Bold(true)

	FocusedLabelStyle = lipgloss.NewStyle().
				Foreground(ColorAccent).
				Bold(true)

	InputContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorderDefault).
				Padding(0, 1).
				MarginBottom(1)

	FocusedInputContainerStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(ColorBorderFocused).
					Padding(0, 1).
					MarginBottom(1)

	// Buttons
	ActiveButtonStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorForeground).
				Background(ColorPrimary).
				Padding(0, 3).
				MarginTop(1)

	InactiveButtonStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Background(lipgloss.Color("#374151")).
				Padding(0, 3).
				MarginTop(1)

	// Feedback & Status Messages
	ErrorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Foreground(ColorError).
			Padding(0, 1).
			MarginTop(1).
			MarginBottom(1)

	SuccessBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Foreground(ColorSuccess).
			Padding(0, 1).
			MarginTop(1).
			MarginBottom(1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorDimmed)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	// Workspace Item Styles
	SelectedItemStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Background(lipgloss.Color("#1E1B4B")).
				Padding(0, 1).
				MarginBottom(1)

	NormalItemStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorderDefault).
			Padding(0, 1).
			MarginBottom(1)

	BadgePublicStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	BadgePrivateStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true)

	BadgeJoinedStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				Bold(true)

	MemberCountStyle = lipgloss.NewStyle().
				Foreground(ColorDimmed).
				Italic(true)
)

// AsciiLogo renders a compact, eye-catching Conclave banner.
const AsciiLogo = `
  ██████╗  ██████╗  ███╗   ██╗  ██████╗ ██╗       █████╗  ██╗   ██╗ ███████╗
 ██╔════╝ ██╔═══██╗ ████╗  ██║ ██╔════╝ ██║      ██╔══██╗ ██║   ██║ ██╔════╝
 ██║      ██║   ██║ ██╔██╗ ██║ ██║      ██║      ███████║ ██║   ██║ █████╗  
 ██║      ██║   ██║ ██║╚██╗██║ ██║      ██║      ██╔══██║ ╚██╗ ██╔╝ ██╔══╝  
 ╚██████╗ ╚██████╔╝ ██║ ╚████║ ╚██████╗ ███████╗ ██║  ██║  ╚████╔╝  ███████╗
  ╚═════╝  ╚═════╝  ╚═╝  ╚═══╝  ╚═════╝ ╚══════╝ ╚═╝  ╚═╝   ╚═══╝   ╚══════╝`

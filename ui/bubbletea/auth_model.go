package bubbletea

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

const (
	focusAuthHost = iota
	focusAuthUser
	focusAuthPass
	focusAuthButton
	focusAuthCount
)

// authModel renders a small DIS credential form and drives Test/Connect flows.
type authModel struct {
	cfg *types.DISUIConfig
	dis types.DISReaderService

	hostInput textinput.Model
	userInput textinput.Model
	passInput textinput.Model

	focused       int
	running       bool
	authenticated bool

	status string
	errMsg string
}

func newAuthModel(cfg *types.DISUIConfig, dis types.DISReaderService) authModel {
	host := textinput.New()
	host.Placeholder = "DIS IP or hostname"
	host.CharLimit = 256
	host.Prompt = "Host: "
	if cfg != nil && cfg.DIS != nil {
		host.SetValue(cfg.DIS.Host)
	}

	user := textinput.New()
	user.Placeholder = "Username"
	user.CharLimit = 64
	user.Prompt = "User: "
	if cfg != nil && cfg.DIS != nil {
		user.SetValue(cfg.DIS.User)
	}

	pass := textinput.New()
	pass.Placeholder = "Password"
	pass.CharLimit = 64
	pass.Prompt = "Pass: "
	pass.EchoMode = textinput.EchoPassword
	pass.EchoCharacter = '•'
	if cfg != nil && cfg.DIS != nil {
		pass.SetValue(cfg.DIS.Password)
	}

	return authModel{
		cfg:       cfg,
		dis:       dis,
		hostInput: host,
		userInput: user,
		passInput: pass,
		focused:   focusAuthHost,
		status:    "Not connected",
	}
}

func (a authModel) Init() (authModel, tea.Cmd) {
	var cmds []tea.Cmd
	if cmd := a.hostInput.Focus(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if a.userInput.Value() != "" && a.passInput.Value() != "" && a.cfg != nil && a.cfg.DIS != nil {
		var authCmd tea.Cmd
		a, authCmd = a.startAuth()
		if authCmd != nil {
			cmds = append(cmds, authCmd)
		}
	}

	if len(cmds) == 0 {
		return a, nil
	}
	return a, tea.Batch(cmds...)
}

func (a authModel) Update(msg tea.Msg, active bool) (authModel, tea.Cmd, bool) {
	var cmds []tea.Cmd
	handled := false

	if active {
		var cmd tea.Cmd
		a.hostInput, cmd = a.hostInput.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		a.userInput, cmd = a.userInput.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		a.passInput, cmd = a.passInput.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	switch m := msg.(type) {
	case tea.KeyMsg:
		if !active {
			break
		}
		if a.running {
			if m.String() == "esc" {
				handled = true
				cmds = append(cmds, dismissAuthModalCmd())
			}
			break
		}
		switch m.String() {
		case "tab":
			handled = true
			a.focused = (a.focused + 1) % focusAuthCount
			var focusCmd tea.Cmd
			a, focusCmd = a.applyFocus()
			if focusCmd != nil {
				cmds = append(cmds, focusCmd)
			}
		case "shift+tab":
			handled = true
			a.focused = (a.focused - 1 + focusAuthCount) % focusAuthCount
			var focusCmd tea.Cmd
			a, focusCmd = a.applyFocus()
			if focusCmd != nil {
				cmds = append(cmds, focusCmd)
			}
		case "ctrl+n":
			handled = true
			a.focused = (a.focused + 1) % focusAuthCount
			var focusCmd tea.Cmd
			a, focusCmd = a.applyFocus()
			if focusCmd != nil {
				cmds = append(cmds, focusCmd)
			}
		case "ctrl+p":
			handled = true
			a.focused = (a.focused - 1 + focusAuthCount) % focusAuthCount
			var focusCmd tea.Cmd
			a, focusCmd = a.applyFocus()
			if focusCmd != nil {
				cmds = append(cmds, focusCmd)
			}
		case "enter":
			if a.focused == focusAuthButton {
				handled = true
				var authCmd tea.Cmd
				a, authCmd = a.startAuth()
				if authCmd != nil {
					cmds = append(cmds, authCmd)
				}
			}
		case "ctrl+a":
			handled = true
			var authCmd tea.Cmd
			a, authCmd = a.startAuth()
			if authCmd != nil {
				cmds = append(cmds, authCmd)
			}
		case "esc":
			handled = true
			cmds = append(cmds, dismissAuthModalCmd())
		}
	case authResultMsg:
		a.running = false
		if m.Err != nil {
			a.errMsg = m.Err.Error()
			a.status = "Connection failed"
			a.authenticated = false
		} else {
			a.errMsg = ""
			a.status = "Connected"
			a.authenticated = true
		}
		cmds = append(cmds, func() tea.Msg { return authStateChangedMsg{authenticated: a.authenticated} })
	}

	if len(cmds) == 0 {
		return a, nil, handled
	}
	return a, tea.Batch(cmds...), handled
}

func (a authModel) View() string {
	parts := []string{
		sectionTitleStyle.Render("DIS Authentication"),
		fieldLabelStyle.Render("Host") + "\n" + a.hostInput.View(),
		fieldLabelStyle.Render("Username") + "\n" + a.userInput.View(),
		fieldLabelStyle.Render("Password") + "\n" + a.passInput.View(),
		a.renderButton(),
		a.renderStatus(),
	}
	parts = append(parts, helpStyle.Render("Ctrl+A or Enter connects • Tab moves • Esc closes"))
	return strings.Join(parts, "\n\n")
}

func (a authModel) renderButton() string {
	style := buttonStyle.Copy()
	label := "Connect"
	if a.running {
		style = style.Copy().Foreground(lipgloss.Color("#777")).Background(lipgloss.Color("#444"))
		label = "Connecting…"
	} else if a.focused == focusAuthButton {
		style = buttonFocusedStyle
	}
	return style.Render(label)
}

func (a authModel) renderStatus() string {
	if a.errMsg != "" {
		return errorStyle.Render(a.errMsg)
	}
	return noticeStyle.Render(a.status)
}

func (a authModel) StatusLine() string {
	if a.authenticated {
		return "DIS Auth: connected"
	}
	if a.running {
		return "DIS Auth: connecting…"
	}
	return "DIS Auth: not connected"
}

func (a authModel) applyFocus() (authModel, tea.Cmd) {
	var cmds []tea.Cmd
	if a.focused == focusAuthHost {
		if cmd := a.hostInput.Focus(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	} else {
		a.hostInput.Blur()
	}

	if a.focused == focusAuthUser {
		if cmd := a.userInput.Focus(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	} else {
		a.userInput.Blur()
	}

	if a.focused == focusAuthPass {
		if cmd := a.passInput.Focus(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	} else {
		a.passInput.Blur()
	}

	if len(cmds) == 0 {
		return a, nil
	}
	return a, tea.Batch(cmds...)
}

func (a authModel) startAuth() (authModel, tea.Cmd) {
	host := strings.TrimSpace(a.hostInput.Value())
	user := strings.TrimSpace(a.userInput.Value())
	pass := strings.TrimSpace(a.passInput.Value())
	if host == "" || user == "" || pass == "" {
		a.errMsg = "Host, user, and password are required"
		a.status = "Missing credentials"
		return a, nil
	}

	if cfg := a.dis.GetConfig(); cfg != nil {
		cfg.Host = host
		cfg.User = user
		cfg.Password = pass
	}

	a.running = true
	a.errMsg = ""
	a.status = "Connecting…"
	return a, authenticateCmd(a.dis, host, user, pass)
}

func authenticateCmd(dis types.DISReaderService, host, user, pass string) tea.Cmd {
	return func() tea.Msg {
		cfg := dis.GetConfig()
		if cfg != nil {
			cfg.Host = host
			cfg.User = user
			cfg.Password = pass
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		if err := dis.TestConnection(ctx); err != nil {
			if err := dis.Connect(ctx); err != nil {
				return authResultMsg{Err: err}
			}
			if err := dis.TestConnection(ctx); err != nil {
				return authResultMsg{Err: err}
			}
		}

		return authResultMsg{Success: true}
	}
}

type authResultMsg struct {
	Success bool
	Err     error
}

type authStateChangedMsg struct {
	authenticated bool
}

type authDismissedMsg struct{}

func dismissAuthModalCmd() tea.Cmd {
	return func() tea.Msg {
		return authDismissedMsg{}
	}
}

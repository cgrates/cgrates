// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/utils"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tuiClient interface {
	Methods() []string
	Describe(string) *utils.MethodDescriptor
	Call(string, any) (any, error)
}

func runTUI(client tuiClient, server string) error {
	_, err := tea.NewProgram(newModel(client, server), tea.WithAltScreen()).Run()
	return err
}

type view uint8

const (
	viewBrowse view = iota
	viewForm
	viewResult
)

var (
	tuiGreen    = lipgloss.Color("114")
	tuiDimGreen = lipgloss.Color("108")
	titleStyle  = lipgloss.NewStyle().Bold(true)
	hintStyle   = lipgloss.NewStyle().Faint(true)
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	selStyle    = lipgloss.NewStyle().Foreground(tuiGreen)
)

type callDoneMsg struct {
	output string
	err    error
}

type model struct {
	client  tuiClient
	view    view
	browser list.Model

	method string
	md     *utils.MethodDescriptor
	forms  []formFrame

	result  viewport.Model
	spin    spinner.Model
	help    help.Model
	calling bool
	status  string
	width   int
	height  int
}

type methodItem struct {
	method      string
	title       string
	description string
}

func (i methodItem) Title() string       { return i.title }
func (i methodItem) Description() string { return i.description }
func (i methodItem) FilterValue() string { return i.method }

func newModel(client tuiClient, server string) model {
	methods := client.Methods()
	items := make([]list.Item, len(methods))
	for i, method := range methods {
		items[i] = methodItem{
			method:      method,
			title:       method,
			description: argSummary(client.Describe(method)),
		}
	}
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(tuiGreen).BorderLeftForeground(tuiGreen)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(tuiDimGreen).BorderLeftForeground(tuiGreen)
	browser := list.New(items, delegate, 0, 0)
	browser.Title = fmt.Sprintf("cgr-console %s", server)
	browser.Styles.Title = browser.Styles.Title.Background(lipgloss.Color("72"))
	browser.Styles.FilterCursor = browser.Styles.FilterCursor.Foreground(tuiGreen)
	browser.SetStatusBarItemName("method", "methods")
	result := viewport.New(0, 0)
	result.SetHorizontalStep(4)
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	return model{client: client, browser: browser, result: result, spin: spin, help: help.New()}
}

func argSummary(md *utils.MethodDescriptor) string {
	if md == nil || len(md.Args) == 0 {
		return "no arguments"
	}
	names := make([]string, len(md.Args))
	for i, field := range md.Args {
		names[i] = field.Name
	}
	return strings.Join(names, ", ")
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.browser.SetSize(msg.Width, msg.Height)
		m.help.Width = msg.Width
		m.result.Width = msg.Width
		m.result.Height = max(msg.Height-2, 1)
		m.result.SetXOffset(0)
		m.result.SetYOffset(m.result.YOffset)
		if len(m.forms) != 0 {
			m.forms[0].value.resize(msg.Width)
		}
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case spinner.TickMsg:
		if !m.calling {
			return m, nil
		}
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	case callDoneMsg:
		m.calling = false
		if msg.err != nil {
			m.status = errStyle.Render("error: " + msg.err.Error())
			return m, nil
		}
		m.result.SetContent(msg.output)
		m.result.GotoTop()
		m.view = viewResult
		return m, nil
	}
	switch m.view {
	case viewForm:
		return m.updateForm(msg)
	case viewResult:
		return m.updateResult(msg)
	default:
		return m.updateBrowse(msg)
	}
}

func (m model) View() string {
	switch m.view {
	case viewForm:
		return m.viewForm()
	case viewResult:
		return titleStyle.Render(m.method) + "\n" + m.result.View() +
			"\n" + m.help.ShortHelpView(helpKeys(
			"arrows", "scroll", "esc", "edit", "q", "methods", "ctrl+c", "quit"))
	default:
		return m.browser.View()
	}
}

func helpKeys(pairs ...string) []key.Binding {
	keys := make([]key.Binding, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		keys = append(keys, key.NewBinding(key.WithKeys(pairs[i]), key.WithHelp(pairs[i], pairs[i+1])))
	}
	return keys
}

func (m model) updateBrowse(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" &&
		m.browser.FilterState() != list.Filtering {
		if item, ok := m.browser.SelectedItem().(methodItem); ok {
			return m.openForm(item.method)
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.browser, cmd = m.browser.Update(msg)
	return m, cmd
}

func (m model) updateResult(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.view = viewForm
			return m, textinput.Blink
		case "q":
			m.view = viewBrowse
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.result, cmd = m.result.Update(msg)
	return m, cmd
}

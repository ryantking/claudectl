package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryantking/agentctl/internal/workspace"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	cleanItemStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	dirtyItemStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

type workspaceItem struct {
	workspace workspace.Workspace
}

func (i workspaceItem) FilterValue() string {
	return i.workspace.Branch
}

func (i workspaceItem) Title() string {
	branch := i.workspace.Branch
	if branch == "" {
		branch = "detached"
	}
	return branch
}

func (i workspaceItem) Description() string {
	isClean, status := i.workspace.IsClean()
	statusText := status
	if isClean {
		statusText = cleanItemStyle.Render("✓ " + status)
	} else {
		statusText = dirtyItemStyle.Render("● " + status)
	}
	return fmt.Sprintf("%s • %s", statusText, i.workspace.Path)
}

type workspacePickerModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m workspacePickerModel) Init() tea.Cmd {
	return nil
}

func (m workspacePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(workspaceItem)
			if ok {
				m.choice = i.workspace.Branch
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m workspacePickerModel) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

func PickWorkspace(workspaces []workspace.Workspace) (string, error) {
	if len(workspaces) == 0 {
		return "", fmt.Errorf("no workspaces available")
	}

	items := make([]list.Item, len(workspaces))
	for i, w := range workspaces {
		items[i] = workspaceItem{workspace: w}
	}

	const defaultWidth = 80
	const listHeight = 14

	l := list.New(items, list.NewDefaultDelegate(), defaultWidth, listHeight)
	l.Title = "Select Workspace"
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = lipgloss.NewStyle()
	l.Styles.HelpStyle = lipgloss.NewStyle()

	m := workspacePickerModel{list: l}

	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	if finalModel.(workspacePickerModel).quitting {
		return "", fmt.Errorf("selection cancelled")
	}

	choice := finalModel.(workspacePickerModel).choice
	if choice == "" {
		return "", fmt.Errorf("no workspace selected")
	}

	return choice, nil
}

// GetWorkspaceArg gets workspace name from args or prompts user to pick one
func GetWorkspaceArg(args []string, workspaces []workspace.Workspace) (string, error) {
	if len(args) > 0 && args[0] != "" {
		return args[0], nil
	}

	// Check if we're in a non-interactive environment
	if !isTerminal(os.Stderr) {
		return "", fmt.Errorf("workspace name required when not in interactive terminal")
	}

	return PickWorkspace(workspaces)
}

func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

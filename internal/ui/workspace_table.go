package ui

import (
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryantking/agentctl/internal/workspace"
)

var (
	tableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Bold(true)
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(true)
	cleanStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	dirtyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

type workspaceTableModel struct {
	table    table.Model
	workspaces []workspace.Workspace
}

func (m workspaceTableModel) Init() tea.Cmd {
	return nil
}

func (m workspaceTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m workspaceTableModel) View() string {
	if len(m.workspaces) == 0 {
		return "\n  No managed workspaces found.\n\n  Create one with: agentctl workspace create <branch>\n\n"
	}
	return "\n" + tableStyle.Render(m.table.View()) + "\n"
}

func ShowWorkspaceTable(workspaces []workspace.Workspace) error {
	columns := []table.Column{
		{Title: "Branch", Width: 30},
		{Title: "Status", Width: 12},
		{Title: "Commit", Width: 10},
		{Title: "Path", Width: 50},
	}

	rows := make([]table.Row, len(workspaces))
	for i, w := range workspaces {
		isClean, status := w.IsClean()
		statusIcon := "✓"
		if !isClean {
			statusIcon = "●"
		}
		statusText := statusIcon + " " + status

		branch := w.Branch
		if branch == "" {
			branch = "detached"
		}

		rows[i] = table.Row{
			branch,
			statusText,
			w.Commit,
			w.Path,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(len(workspaces)+2, 20)),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = selectedStyle
	s.Cell = lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)
	t.SetStyles(s)

	m := workspaceTableModel{
		table:      t,
		workspaces: workspaces,
	}

	if _, err := tea.NewProgram(m, tea.WithOutput(os.Stderr)).Run(); err != nil {
		return err
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

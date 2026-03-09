package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

type statusModel struct {
	workflows []*workflow.Workflow
}

func (m statusModel) Init() tea.Cmd {
	return nil
}

func (m statusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		return m, tea.Quit
	}
	return m, nil
}

func (m statusModel) View() string {
	sep := "\n" + ui.StyleMuted.Render(strings.Repeat("─", 32)) + "\n"
	views := make([]string, 0, len(m.workflows))
	for _, wf := range m.workflows {
		views = append(views, ui.RenderStatus(wf))
	}
	body := strings.Join(views, sep)
	hint := "\n" + ui.StyleMuted.Render("press any key to exit")
	return "\n" + body + hint + "\n"
}

func runStatusModel(wfs []*workflow.Workflow) error {
	p := tea.NewProgram(statusModel{workflows: wfs})
	_, err := p.Run()
	return err
}

package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"

	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

var statusInput = os.Stdin
var statusOutput = os.Stdout

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
	body := renderStatusBody(m.workflows)
	hint := "\n" + ui.StyleMuted.Render("press any key to exit")
	return "\n" + body + hint + "\n"
}

func runStatusModel(wfs []*workflow.Workflow) error {
	if !hasTTY(statusInput) || !hasTTY(statusOutput) {
		fmt.Fprintln(statusOutput, renderStatusBody(wfs))
		return nil
	}
	p := tea.NewProgram(statusModel{workflows: wfs})
	_, err := p.Run()
	return err
}

func renderStatusBody(wfs []*workflow.Workflow) string {
	sep := "\n" + ui.StyleMuted.Render(strings.Repeat("─", 32)) + "\n"
	views := make([]string, 0, len(wfs))
	for _, wf := range wfs {
		views = append(views, ui.RenderStatus(wf))
	}
	return strings.Join(views, sep)
}

func hasTTY(f *os.File) bool {
	if f == nil {
		return false
	}
	fd := f.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

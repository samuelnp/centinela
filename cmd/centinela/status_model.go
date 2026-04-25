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
var statusHasTTY = hasTTY
var runInteractiveStatus = func(wfs []*workflow.Workflow) error {
	p := tea.NewProgram(statusModel{workflows: wfs})
	_, err := p.Run()
	return err
}

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
	if !statusHasTTY(statusInput) || !statusHasTTY(statusOutput) {
		fmt.Fprintln(statusOutput, renderStatusBody(wfs))
		return nil
	}
	return runInteractiveStatus(wfs)
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

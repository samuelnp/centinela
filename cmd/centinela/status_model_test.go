package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestStatusModelMethods(t *testing.T) {
	wf := &workflow.Workflow{Feature: "f", CurrentStep: "plan", Steps: map[string]workflow.StepState{"plan": {Status: "in-progress"}}}
	m := statusModel{workflows: []*workflow.Workflow{wf}}
	if m.Init() != nil {
		t.Fatal("Init should return nil cmd")
	}
	if v := m.View(); v == "" {
		t.Fatal("View should render text")
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Update should return quit cmd on key")
	}
	_, cmd2 := m.Update(struct{}{})
	if cmd2 != nil {
		t.Fatal("Update should return nil cmd on non-key msg")
	}
}

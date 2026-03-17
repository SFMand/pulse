package config

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

type StatusUpdate struct {
	Name      string
	State     string
	Detail    string
	Latency   time.Duration
	LastCheck time.Time
}

type statusMsg StatusUpdate
type streamClosedMsg struct{}
type doneMsg struct{}

type monitorModel struct {
	serviceOrder []string
	statuses     map[string]StatusUpdate
	updates      <-chan StatusUpdate
	done         <-chan struct{}
}

func InitializeModel(targets []Target, updates <-chan StatusUpdate, done <-chan struct{}) monitorModel {
	order := make([]string, 0, len(targets))
	statuses := make(map[string]StatusUpdate, len(targets))

	for _, target := range targets {
		order = append(order, target.Name)
		statuses[target.Name] = StatusUpdate{Name: target.Name, State: "PENDING"}
	}

	return monitorModel{
		serviceOrder: order,
		statuses:     statuses,
		updates:      updates,
		done:         done,
	}
}

func waitForStatusUpdate(updates <-chan StatusUpdate) tea.Cmd {
	return func() tea.Msg {
		update, ok := <-updates
		if !ok {
			return streamClosedMsg{}
		}
		return statusMsg(update)
	}
}

func waitForDone(done <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		<-done
		return doneMsg{}
	}
}

func (m monitorModel) Init() tea.Cmd {
	return tea.Batch(waitForStatusUpdate(m.updates), waitForDone(m.done))
}

func (m monitorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statusMsg:
		update := StatusUpdate(msg)
		m.statuses[update.Name] = update
		return m, waitForStatusUpdate(m.updates)
	case doneMsg, streamClosedMsg:
		return m, tea.Quit
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func formatLatency(latency time.Duration) string {
	if latency <= 0 {
		return "-"
	}
	return latency.String()
}

func (m monitorModel) View() tea.View {
	var b strings.Builder
	b.WriteString("Pulse Monitor\n\n")
	b.WriteString(fmt.Sprintf("%-24s %-10s %-12s\n", "Name", "State", "Latency"))

	for _, name := range m.serviceOrder {
		status := m.statuses[name]
		b.WriteString(fmt.Sprintf("%-24s %-10s %-12s\n", status.Name, status.State, formatLatency(status.Latency)))
	}

	b.WriteString("\nCtrl+C to quit\n")

	return tea.NewView(b.String())
}

package config

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

type myModel struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func InitializeModel() myModel {
	return myModel{
		choices:  []string{"Target 1", "Target 2", "Target 3"},
		selected: make(map[int]struct{}),
	}
}

func (m myModel) Init() tea.Cmd {
	return nil
}

func (m myModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", "space":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}
	return m, nil
}

func (m myModel) View() tea.View {
	s := "Pick a Target\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = "->"
		}
		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "X"
		}
		s += fmt.Sprintf("%s [%s] %v\n", cursor, checked, choice)
	}

	s += "Press q to quit.\n"

	return tea.NewView(s)
}

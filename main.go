package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/santiagomac/springboot-init-v2/utils"
)

type model struct {
	textInput    textinput.Model
	step         int
	projectName  string
	groupID      string
	artifactID   string
	dependencies []string
	options      []string
	cursor       int
	selected     map[int]struct{}
	progress     progress.Model
	progressCh   chan int
	doneCh       chan struct{}
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Your project name"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	progressModel := progress.New()
	progressModel.SetPercent(0)

	return model{
		step:       0,
		textInput:  ti,
		options:    []string{"web", "webflux", "jpa", "security", "thymeleaf", "actuator"},
		selected:   map[int]struct{}{},
		progress:   progressModel,
		progressCh: make(chan int),
		doneCh:     make(chan struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if m.step == 3 {
				return m, tea.Quit
			}
			m.step++
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
				m.dependencies = append(m.dependencies, m.options[m.cursor])
			}
		}
	case string:
		if msg == "project created" {
			return m, tea.Quit
		}
		cmd := m.progress.IncrPercent(0.25)
		project := &utils.SpringProject{
			GroupID:      m.groupID,
			ArtifactID:   m.artifactID,
			ProjectName:  m.projectName,
			Dependencies: m.dependencies,
		}
		return m, tea.Batch(cmd, utils.CreateProject(project, &m.progress))
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

	return m, cmd
}

func (m model) View() string {
	s := ""

	switch m.step {
	case 0:
		s += "Welcome to Spring Boot CLI Init\n\n"
		s += fmt.Sprintf("%s\n\n%s", m.textInput.View(), "(esc or q to quit")
	case 1:
		s += "Select the dependencies for your project: \n\n"

		for i, choice := range m.options {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			checked := " "
			if _, ok := m.selected[i]; ok {
				checked = "x"
			}

			s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
		}
	case 2:
		pad := strings.Repeat(" ", 2)
		s += "\n" + pad + m.progress.View() + "\n\n"
		fmt.Println("CASE 2")
		m.progress.View()
	}
	s += "\n Press 'q' to exit"
	return s
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Printf("Error to init the program: %v\n", err)
	}
}

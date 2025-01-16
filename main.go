package main

import (
	"fmt"
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

type tickMsg string

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Your project name"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	progressModel := progress.New(progress.WithDefaultGradient())
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
			if m.step == 0 {
				m.projectName = m.textInput.Value()
			}
			m.step++
			if m.step == 2 {
				project := &utils.SpringProject{
					GroupID:      m.groupID,
					ArtifactID:   m.artifactID,
					ProjectName:  m.projectName,
					Dependencies: m.dependencies,
				}
				return m, firstMethod(project)
			}
			if m.step == 3 {
				return m, tea.Quit
			}
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
	case tickMsg:
		if string(msg) == "init" {
			cmd := m.progress.IncrPercent(0.10)
			return m, cmd
		}

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
		s += "Creating project...\n\n"
		s += fmt.Sprintf("%s\n", m.progress.View())
		s += fmt.Sprintf("Progress: %.0f%%\n", m.progress.Percent()*100)
	}
	s += "\n Press 'q' to exit"
	return s
}

func progressCmd(t string) tea.Cmd {
	return func() tea.Msg {
		return tickMsg(t)
	}
}

func firstMethod(project *utils.SpringProject) tea.Cmd {
	result := 2 + 2
	fmt.Printf("Result: %d", result)
	data, projectName, err := utils.CreateProject(project)
	if err != nil {
		tea.Quit()
	}

	err = utils.Unzip(data, projectName)
	if err != nil {
		tea.Quit()
	}
	return progressCmd("init")
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Printf("Error to init the program: %v\n", err)
	}
}

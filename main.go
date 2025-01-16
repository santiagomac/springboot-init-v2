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
	creationStep int
}

type tickMsg int

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Your project name"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	progressModel := progress.New(progress.WithDefaultGradient())
	progressModel.SetPercent(0)

	return model{
		step:         0,
		textInput:    ti,
		options:      []string{"web", "webflux", "jpa", "security", "thymeleaf", "actuator"},
		selected:     map[int]struct{}{},
		progress:     progressModel,
		creationStep: 1,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

var (
	data        *[]byte
	projectName *string
	cmdCreation tea.Cmd
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.step == 3 {
		return m, tea.Quit
	}
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
			if m.step == 2 && m.creationStep == 1 {
				project := &utils.SpringProject{
					GroupID:      m.groupID,
					ArtifactID:   m.artifactID,
					ProjectName:  m.projectName,
					Dependencies: m.dependencies,
				}
				data, projectName, cmdCreation = downloadProject(&m, project)
				m.creationStep++
				return m, cmdCreation
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
		if int(msg) == 1 {
			cmd := m.progress.IncrPercent(0.5)
			return m, tea.Batch(cmd)
		}
		if int(msg) == 2 {
			cmd := m.progress.IncrPercent(0.5)
			m.step++
			return m, cmd
		}
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	default:
		if m.step == 2 && m.creationStep == 2 {
			_, err := unzip(data, projectName)
			if err != nil {
				return m, tea.Quit
			}
			return m, progressCmd(2)
		}
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
		s += fmt.Sprintf("\n\nProgess: %v\n", m.progress.Percent())
		return s
	case 3:
		s += fmt.Sprintf("\n\n Project '%s' created successfully\n\n", m.projectName)
		return s
	}
	s += "\n Press 'q' to exit"
	return s
}

func progressCmd(creationStep int) tea.Cmd {
	return func() tea.Msg {
		return tickMsg(creationStep)
	}
}

func downloadProject(m *model, project *utils.SpringProject) (dataReturn *[]byte, projectNameReturn *string, cmd tea.Cmd) {
	data, projectName, err := utils.CreateProject(project)
	if err != nil {
		return nil, nil, tea.Quit
	}

	return &data, &projectName, progressCmd(m.creationStep)
}

func unzip(data *[]byte, dest *string) (tea.Cmd, error) {
	err := utils.Unzip(*data, *dest)
	if err != nil {
		return nil, err
	}

	return progressCmd(2), nil
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Printf("Error to init the program: %v\n", err)
	}
}

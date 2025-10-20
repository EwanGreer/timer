package commands

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/beeep"
)

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type StartModel struct {
	Remaining time.Duration
	width     int
	height    int
	running   bool
	done      bool
}

func (m StartModel) Init() tea.Cmd {
	return tick()
}

func (m StartModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		if m.Remaining <= 0 {
			m.done = true
			return m, nil
		}
		m.Remaining -= time.Second
		return m, tick()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m StartModel) View() string {
	if m.done {
		beeep.Notify("Your Timer is Complete!", "Your timer is completed!", "")

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("10")).
				Render(doneArt),
		)
	}

	mins := int(m.Remaining.Minutes())
	secs := int(m.Remaining.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", mins, secs)

	bigStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Render(renderBigClock(timeStr))

	if m.Remaining < time.Second*60 {
		bigStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("13")).
			Render(renderBigClock(timeStr))
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bigStyle)
}

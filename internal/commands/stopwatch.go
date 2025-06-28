package commands

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StopWatchModel struct {
	width              int
	height             int
	StartTime          time.Time
	paused             bool
	elapsedBeforePause time.Duration
}

func (m StopWatchModel) Init() tea.Cmd {
	m.StartTime = time.Now()
	return tick()
}

func (m StopWatchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		if m.paused {
			return m, tick()
		}
		return m, tick()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "p", " ":
			if m.paused {
				m.StartTime = time.Now().Add(-m.elapsedBeforePause)
				m.paused = false
			} else {
				m.elapsedBeforePause = time.Since(m.StartTime)
				m.paused = true
			}
		}
	}
	return m, nil
}

func (m StopWatchModel) View() string {
	color := lipgloss.Color("12")
	var elapsed time.Duration
	if m.paused {
		elapsed = m.elapsedBeforePause
		color = lipgloss.Color("13")
	} else {
		elapsed = time.Since(m.StartTime)
	}

	mins := int(elapsed.Minutes())
	secs := int(elapsed.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", mins, secs)

	bigStyle := lipgloss.NewStyle().
		Foreground(color).
		Render(renderBigClock(timeStr))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bigStyle)
}

package commands

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/beeep"
)

type tickMsg time.Time

// notify is the desktop notification function. It is a variable so tests can
// stub it.
var notify = beeep.Notify

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// notifyDone sends the completion notification once.
func (m StartModel) notifyDone() tea.Cmd {
	return func() tea.Msg {
		message := "Your timer is completed!"
		if m.Name != "" {
			message = fmt.Sprintf(`Your timer "%s" is completed!`, m.Name)
		}
		notify("Your Timer is Complete!", message, "")
		return nil
	}
}

type StartModel struct {
	Remaining time.Duration
	Name      string
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
			return m, m.notifyDone()
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

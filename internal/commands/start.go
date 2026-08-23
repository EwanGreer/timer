package commands

import (
	"fmt"
	"time"

	"github.com/EwanGreer/timer/internal/art"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/beeep"
)

type tickMsg time.Time

// notify is a variable so tests can stub it.
var notify = beeep.Notify

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func NotifyComplete(name string) error {
	message := "Your timer is completed!"
	if name != "" {
		message = fmt.Sprintf(`Your timer "%s" is completed!`, name)
	}
	return notify("Your Timer is Complete!", message, "")
}

func (m StartModel) notifyDone() tea.Cmd {
	return func() tea.Msg {
		NotifyComplete(m.Name)
		return nil
	}
}

type StartModel struct {
	Remaining time.Duration
	Name      string
	Art       *art.Set
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
	s := artFor(m.Art)

	if m.done {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("10")).
				Render(s.Done()),
		)
	}

	mins := int(m.Remaining.Minutes())
	secs := int(m.Remaining.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", mins, secs)

	bigStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Render(s.RenderClock(timeStr))

	if m.Remaining < time.Second*60 {
		bigStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("13")).
			Render(s.RenderClock(timeStr))
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bigStyle)
}

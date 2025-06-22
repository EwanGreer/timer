package commands

import (
	"fmt"
	"strconv"

	"github.com/EwanGreer/timer-cli/internal/deps"
	"github.com/EwanGreer/timer-cli/internal/models"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type CobraCmdFunc = func(cmd *cobra.Command, args []string)

type Repository interface {
	GetAll() ([]models.Timer, error)
}

func List(GetDeps func() *deps.Deps) CobraCmdFunc {
	return func(cmd *cobra.Command, args []string) {
		items, err := GetDeps().DB.GetAll()
		cobra.CheckErr(err)

		columns := []table.Column{
			{Title: "ID", Width: 4},
			{Title: "Name", Width: 10},
		}

		rows := []table.Row{}
		for _, item := range items {
			rows = append(rows, []string{strconv.Itoa(int(item.ID)), item.GetName()})
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithHeight(7),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		s.Selected = lipgloss.Style{}

		t.SetStyles(s)

		baseStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

		fmt.Println(baseStyle.Render(t.View()))
	}
}

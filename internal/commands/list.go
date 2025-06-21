package commands

import (
	"fmt"
	"log/slog"

	"github.com/EwanGreer/timer-cli/internal/models"
	"github.com/spf13/cobra"
)

type CobraCmdFunc = func(cmd *cobra.Command, args []string)

type Repository interface {
	GetAll() ([]models.Timer, error)
}

func List(repo Repository) CobraCmdFunc {
	return func(cmd *cobra.Command, args []string) {
		slog.Debug("LIST CALLED:")
		items, err := repo.GetAll()
		cobra.CheckErr(err)

		for _, item := range items {
			fmt.Println(item.GetName())
		}
	}
}

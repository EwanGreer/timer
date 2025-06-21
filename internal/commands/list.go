package commands

import (
	"fmt"

	"github.com/EwanGreer/timer-cli/internal/deps"
	"github.com/EwanGreer/timer-cli/internal/models"
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

		for _, item := range items {
			fmt.Println(item.GetName())
		}
	}
}

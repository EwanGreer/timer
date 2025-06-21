package deps

import (
	"github.com/EwanGreer/timer-cli/internal/repository"
)

var dependencies = &Deps{}

type Deps struct {
	DB *repository.SQLiteDatabase
}

func NewDeps(DB *repository.SQLiteDatabase) {
	dependencies = &Deps{
		DB: DB,
	}
}

func GetDeps() *Deps {
	return dependencies
}

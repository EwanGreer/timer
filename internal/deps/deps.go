package deps

import "github.com/EwanGreer/timer-cli/internal/repository"

var Dependencies = &Deps{}

type Deps struct {
	DB *repository.SQLiteDatabase
}

func NewDeps(DB *repository.SQLiteDatabase) {
	Dependencies = &Deps{
		DB: DB,
	}
}

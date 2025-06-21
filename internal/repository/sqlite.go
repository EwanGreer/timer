package repository

import (
	"github.com/EwanGreer/timer-cli/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLiteDatabase struct {
	*gorm.DB
}

func NewSqliteDatabase(path string) *SQLiteDatabase {
	var err error
	DB, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil
	}

	return &SQLiteDatabase{DB: DB}
}

func (s *SQLiteDatabase) GetAll() ([]models.Timer, error) {
	items := []models.Timer{
		{
			Model: gorm.Model{ID: 1},
			Name:  "Item 1",
		},
		{
			Model: gorm.Model{ID: 2},
			Name:  "",
		},
	}

	return items, nil
}

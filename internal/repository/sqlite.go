package repository

import (
	"github.com/EwanGreer/timer-cli/internal/models"
	"github.com/spf13/cobra"
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

	DB.AutoMigrate(&models.Timer{})

	return &SQLiteDatabase{DB: DB}
}

func (s *SQLiteDatabase) GetAll() ([]models.Timer, error) {
	items := []models.Timer{}
	tx := s.Find(&items)
	cobra.CheckErr(tx.Error)

	return items, nil
}

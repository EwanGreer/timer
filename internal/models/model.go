package models

import (
	"strconv"

	"gorm.io/gorm"
)

type Storable interface {
	GetName() string
}

type Timer struct {
	gorm.Model
	Name string `json:"name"`
}

func (t *Timer) GetName() string {
	if t.Name == "" {
		return strconv.Itoa(int(t.ID))
	}
	return t.Name
}

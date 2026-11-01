package database

import (
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func Connect(db *gorm.DB) Handler {
	return Handler{db}
}

package models

import "gorm.io/gorm"

// Note - stores note
type Note struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
}

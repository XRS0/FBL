package models

import "gorm.io/gorm"

type Handler struct {
	DB *gorm.DB
}

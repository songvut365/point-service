package model

import "gorm.io/gorm"

type Point struct {
	gorm.Model
	Level     string
	Remaining uint
}


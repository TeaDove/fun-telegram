package db_repository

import (
	"time"
)

type WithId struct {
	ID uint `gorm:"primarykey"`
}

type WithCreatedAt struct {
	CreatedAt time.Time
}

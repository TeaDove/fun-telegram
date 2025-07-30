package db_repository

import (
	"time"
)

type WithID struct {
	ID uint `gorm:"primarykey"`
}

type WithCreatedAt struct {
	CreatedAt time.Time `gorm:"index"`
}

type WithCreatedInDBAt struct {
	CreatedInDBAt time.Time `gorm:"index"`
}

type WithUpdatedAt struct {
	UpdatedAt time.Time `gorm:"index"`
}

type WithUpdatedInDBAt struct {
	UpdatedInDBAt time.Time `gorm:"index"`
}

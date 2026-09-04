package model

import "time"

type URL struct {
	ID        uint   `gorm:"primaryKey"`
	Code      string `gorm:"uniqueIndex;size:16;not null"`
	LongURL   string `gorm:"not null"`
	CreatedAt time.Time
}

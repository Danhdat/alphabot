package models

import "time"
type NotificationLog struct {
	ID        uint      `gorm:"primaryKey"`
	Symbol    string    `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"not null"`
	Direction int       // "1:bullish", "2:bearish", "0:none"
}

func (NotificationLog) TableName() string {
	return "notification_logs"
}

type HolderHistory struct {
	ID           uint           `gorm:"primaryKey"`
	Symbol       string         `gorm:"not null;index"`
	Holders      int            `gorm:"not null"`
	ChangeAmount float64        `gorm:"not null"`
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
}

func (HolderHistory) TableName() string {
	return "holder_history"
}
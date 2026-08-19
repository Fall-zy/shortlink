package model

import "time"

type AccessLog struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ShortCode  string    `gorm:"index;size:20;not null" json:"short_code"`
	IP         string    `gorm:"size:45" json:"ip"`
	UserAgent  string    `gorm:"size:500" json:"user_agent"`
	Referer    string    `gorm:"size:500" json:"referer"`
	AccessTime time.Time `gorm:"index" json:"access_time"`
}

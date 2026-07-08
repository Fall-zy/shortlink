package model

import "time"

type ShortLink struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	ShortCode   string     `gorm:"uniqueIndex;size:10;not null" json:"short_code"`
	OriginalURL string     `gorm:"type:text;not null" json:"original_url"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsDeleted   int8       `gorm:"default:0" json:"is_deleted"`
}

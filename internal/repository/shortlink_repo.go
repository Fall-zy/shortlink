package repository

import (
	"shortlink/internal/model"

	"gorm.io/gorm"
)

type ShortLinkRepo struct {
	db *gorm.DB
}

func NewShortLinkRepo(db *gorm.DB) *ShortLinkRepo {
	return &ShortLinkRepo{db: db}
}

func (r *ShortLinkRepo) Create(link *model.ShortLink) error {
	return r.db.Create(link).Error
}

func (r *ShortLinkRepo) GetByShortCode(code string) (*model.ShortLink, error) {
	var link model.ShortLink
	err := r.db.Where("short_code = ? AND is_deleted = 0", code).First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

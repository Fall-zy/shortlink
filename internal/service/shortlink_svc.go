package service

import (
	"errors"
	"shortlink/internal/model"
	"shortlink/internal/repository"
	"shortlink/internal/utils"

	"gorm.io/gorm"
)

type ShortLinkSvc struct {
	repo *repository.ShortLinkRepo
}

func NewShortLinkSvc(repo *repository.ShortLinkRepo) *ShortLinkSvc {
	return &ShortLinkSvc{repo: repo}
}

func (s *ShortLinkSvc) CreateShortLink(originalURL string) (*model.ShortLink, error) {
	//1.生成雪花ID
	id, err := utils.Flake.NextID()
	if err != nil {
		return nil, err
	}
	//2.生成短码
	code := utils.EncodeBase62(id)
	//3.构造完整记录
	link := &model.ShortLink{
		ID:          id,
		ShortCode:   code,
		OriginalURL: originalURL,
	}
	//4.一次性插入
	err = s.repo.Create(link)
	if err != nil {
		return nil, err
	}
	return link, nil
}

func (s *ShortLinkSvc) GetOriginalURL(shortCode string) (string, error) {
	link, err := s.repo.GetByShortCode(shortCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("短链接不存在")
		}
		return "", err
	}
	return link.OriginalURL, nil
}

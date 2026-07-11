package service

import (
	"context"
	"errors"
	"shortlink/internal/model"
	"shortlink/internal/repository"
	"shortlink/internal/utils"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	cacheKeyPrefix     = "shortlink:"
	invalidPlaceholder = "__INVALID__"
	invalidTTL         = 1 * time.Minute
	normalTTL          = 1 * time.Hour
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

	ctx := context.Background()
	utils.Rdb.Set(ctx, cacheKeyPrefix+code, originalURL, normalTTL)
	utils.Logger.Info("短链接已创建并缓存", zap.String("code", code))

	return link, nil
}

func (s *ShortLinkSvc) GetOriginalURL(shortCode string) (string, error) {
	ctx := context.Background()
	key := cacheKeyPrefix + shortCode

	val, err := utils.Rdb.Get(ctx, key).Result()
	if err == nil {
		if val == invalidPlaceholder {
			return "", errors.New("短链接不存在")
		}
		utils.Logger.Debug("短链接缓存命中", zap.String("code", shortCode))
		return val, nil
	}
	if !errors.Is(err, redis.Nil) {
		utils.Logger.Warn("Redis 查询异常", zap.String("code", shortCode), zap.Error(err))
		//继续查数据库
	}

	link, err := s.repo.GetByShortCode(shortCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			//防止缓存穿透
			utils.Rdb.Set(ctx, key, invalidPlaceholder, invalidTTL)
			return "", errors.New("短链接不存在")
		}
		return "", err
	}

	utils.Rdb.Set(ctx, key, link.OriginalURL, normalTTL)
	utils.Logger.Info("短链接已缓存", zap.String("code", shortCode))

	return link.OriginalURL, nil
}

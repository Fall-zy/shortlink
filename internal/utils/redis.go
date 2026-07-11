package utils

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"time"
)

var Rdb *redis.Client

func InitRedis(cfg RedisCfg) {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Rdb.Ping(ctx).Err(); err != nil {
		Logger.Fatal("Redis 连接失败", zap.Error(err))
	}
	Logger.Info("Redis 连接成功", zap.String("addr", cfg.Addr))
}

func CloseRedis() {
	if Rdb != nil {
		if err := Rdb.Close(); err != nil {
			Logger.Error("关闭 Redis 连接失败", zap.Error(err))
		}
	}
}

type RedisCfg struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

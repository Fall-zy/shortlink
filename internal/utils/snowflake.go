package utils

import (
	"log"

	"github.com/sony/sonyflake"
)

var Flake *sonyflake.Sonyflake

func InitSnowflake() {
	//单机部署可用空Settings,内部会用随机数生成机器ID
	Flake = sonyflake.NewSonyflake(sonyflake.Settings{})
	if Flake == nil {
		log.Fatal("sonyflake初始化失败")
	}
}

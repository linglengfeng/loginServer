package db_redis

import (
	"context"
	"loginServer/config"
	inredis "loginServer/pkg/redis"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var DB *redis.Client

func Start() {
	redisip := config.Config.GetString("redis.ip")
	if redisip != "" {
		redisip := config.Config.GetString("redis.ip")
		redislink := inredis.Link{
			Password: config.Config.GetString("redis.password"),
			Db:       config.Config.GetInt("redis.db"),
			Ip:       redisip,
			Port:     config.Config.GetString("redis.port"),
		}
		DB = inredis.Start(redislink)
	}
}

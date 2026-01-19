package redis

import (
	"github.com/redis/go-redis/v9"
)

type Link struct {
	User     string
	Password string
	Db       int
	Ip       string
	Port     string
}

func Start(redislink Link) *redis.Client {
	user := redislink.User
	password := redislink.Password
	db := redislink.Db
	addr := redislink.Ip + ":" + redislink.Port
	dbcli := redis.NewClient(&redis.Options{
		Username: user,
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return dbcli
}

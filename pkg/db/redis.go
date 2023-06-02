package db

import "github.com/WuKongIM/WuKongChatServer/pkg/redis"

func NewRedis(addr string) *redis.Conn {
	return redis.New(addr)
}

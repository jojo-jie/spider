package cache

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"spider/internal/conf"
)

var ProviderSet = wire.NewSet(NewCache)

// Cache .
type Cache struct {
	// TODO wrapped cache client example redis or memcache
	client *redis.Client
}

func NewCache(c *conf.Data, logger log.Logger) (*Cache, func(), error) {
	client := redis.NewClient(&redis.Options{
		Addr:     c.Redis.GetAddr(),
		Password: "",
		DB:       0,
	})
	cleanup := func() {
		log.NewHelper(logger).Info("closing the cache resources")
		client.Close()
	}
	return &Cache{}, cleanup, nil
}

package biz

import (
	"github.com/google/wire"
	"spider/internal/cache"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(cache.NewCache, NewGreeterUsecase)

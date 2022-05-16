//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"spider/internal/biz"
	"spider/internal/conf"
	"spider/internal/data"
	"spider/internal/server"
	"spider/internal/service"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger, string) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}

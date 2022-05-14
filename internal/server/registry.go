package server

import (
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"spider/internal/conf"
)

// NewRegistry 服务注册发现
func NewRegistry(c *conf.Data, logger log.Logger) (registry.Registrar, func(), error) {
	switch c.GetRegistryType().String() {
	case conf.RegistryType_name[int32(conf.RegistryType_ETCD)]:
		client, err := clientv3.New(clientv3.Config{
			Endpoints: c.Etcd.GetAddr(),
		})
		cleanup := func() {
			log.NewHelper(logger).Info("closing the etcd registry resources")
			client.Close()
		}
		return etcd.New(client), cleanup, errors.Wrap(err, "registry etcd")
	case conf.RegistryType_name[int32(conf.RegistryType_CONSUL)]:
		client, err := api.NewClient(api.DefaultConfig())
		cleanup := func() {
			log.NewHelper(logger).Info("closing the consul registry resources")
		}
		return consul.New(client), cleanup, errors.Wrap(err, "registry consul")
	}
	return nil, func() {
		log.NewHelper(logger).Info("local config file")
	}, nil
}

package data

import (
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/hashicorp/consul/api"
	clientv3 "go.etcd.io/etcd/client/v3"
	"spider/internal/conf"
)

// NewRegistry 服务注册发现
func NewRegistry(c *conf.Data, data *Data, logger log.Logger) (reg registry.Registrar, err error) {
	switch c.GetRegistryType().String() {
	case conf.RegistryType_name[int32(conf.RegistryType_ETCD)]:
		client, err := clientv3.New(clientv3.Config{
			Endpoints: c.Etcd.GetAddr(),
		})
		return etcd.New(client), err
	case conf.RegistryType_name[int32(conf.RegistryType_CONSUL)]:
		client, err := api.NewClient(api.DefaultConfig())
		return consul.New(client), err
	default:
		return nil, nil
	}
}

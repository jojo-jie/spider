package data

import (
	"context"
	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"spider/ent"
	"spider/internal/conf"
	"time"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	// TODO wrapped database client
	client *ent.Client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	drv, err := sql.Open(c.Database.GetDriver(), c.Database.GetSource())
	if err != nil {
		return nil, nil, errors.Wrap(err, "sql open fail")
	}
	// 获取数据库驱动中的sql.DB对象
	db := drv.DB()
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)
	client := ent.NewClient(ent.Driver(drv))
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		client.Close()
		db.Close()
	}
	if err := client.Schema.Create(context.Background()); err != nil {
		return nil, nil, errors.Wrap(err, "client Schema fail")
	}
	return &Data{client: client}, cleanup, nil
}

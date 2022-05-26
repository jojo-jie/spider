package data

import (
	"ariga.io/sqlcomment"
	"context"
	"entgo.io/ent/dialect/sql"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

type CustomCommenter struct{}

func (mcc CustomCommenter) Tag(ctx context.Context) sqlcomment.Tags {
	return sqlcomment.Tags{
		"key": "value",
	}
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	drv, err := sql.Open(c.Database.GetDriver(), c.Database.GetSource())
	if err != nil {
		return nil, nil, errors.Wrap(err, "sql open fail")
	}
	// 获取数据库驱动中的sql.DB对象
	db := drv.DB()
	db.SetMaxIdleConns(int(c.GetDatabase().GetMaxIdleConns()))
	db.SetMaxOpenConns(int(c.GetDatabase().GetMaxOpenConns()))
	db.SetConnMaxLifetime(time.Duration(c.GetDatabase().GetConnMaxLifetime().Seconds))

	client := ent.NewClient(ent.Driver(drv))
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		client.Close()
		db.Close()
	}
	if err := client.Schema.Create(context.Background()); err != nil {
		return nil, nil, errors.Wrap(err, "client Schema fail")
	}

	// Add a global hook that runs on all types and all operations.
	client.Debug().Use(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			start := time.Now()
			defer func() {
				msg := fmt.Sprintf("Op=%s\tType=%s\tTime=%s\tConcreteType=%T\n", m.Op(), m.Type(), time.Since(start), m)
				log.NewHelper(logger).Infof(msg)
				_, span := otel.GetTracerProvider().Tracer("").Start(ctx, "ent sql")
				span.SetAttributes(attribute.String("sql_msg", msg))
				span.SetName("zzzyf")
				span.SetStatus(codes.Unset, "ok")
				defer span.End()
			}()
			return next.Mutate(ctx, m)
		})
	})
	return &Data{client: client}, cleanup, nil
}

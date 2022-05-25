package data

import (
	"ariga.io/sqlcomment"
	"context"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"spider/ent"
	"spider/internal/conf"
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
	defer func() {
		if tp, ok := otel.GetTracerProvider().(*tracesdk.TracerProvider); ok {
			tp.Shutdown(context.Background())
		}
	}()
	drv, err := sql.Open(c.Database.GetDriver(), c.Database.GetSource())
	if err != nil {
		return nil, nil, errors.Wrap(err, "sql open fail")
	}
	// 获取数据库驱动中的sql.DB对象
	/*db := drv.DB()
	db.SetMaxIdleConns(int(c.GetDatabase().GetMaxIdleConns()))
	db.SetMaxOpenConns(int(c.GetDatabase().GetMaxOpenConns()))
	db.SetConnMaxLifetime(time.Duration(c.GetDatabase().GetConnMaxLifetime().Seconds))
	*/

	commentedDriver := sqlcomment.NewDriver(dialect.Debug(drv),
		sqlcomment.WithTagger(
			// add tracing info with Open Telemetry.
			sqlcomment.NewOTELTagger(),
			// use your custom commenter

		),
		// add `db_driver` version tag
		sqlcomment.WithDriverVerTag(),
		// add some global tags to all queries
		sqlcomment.WithTags(sqlcomment.Tags{
			sqlcomment.KeyApplication: "bootcamp",
			sqlcomment.KeyFramework:   "go-chi",
		}))

	client := ent.NewClient(ent.Driver(commentedDriver))

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		client.Close()
		//db.Close()
	}
	if err := client.Schema.Create(context.Background()); err != nil {
		return nil, nil, errors.Wrap(err, "client Schema fail")
	}
	// Add a global hook that runs on all types and all operations.
	/*client.Debug().Use(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			start := time.Now()
			defer func() {
				log.NewHelper(logger).Infof("Op=%s\tType=%s\tTime=%s\tConcreteType=%T\n", m.Op(), m.Type(), time.Since(start), m)
			}()
			return next.Mutate(ctx, m)
		})
	})*/
	return &Data{client: client}, cleanup, nil
}

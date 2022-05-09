package main

import (
	"context"
	"flag"
	"github.com/go-kratos/kratos/v2"
	"io"
	"os"
	"path/filepath"
	"time"

	cfg "github.com/go-kratos/kratos/contrib/config/etcd/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	clientv3 "go.etcd.io/etcd/client/v3"
	google_grpc "google.golang.org/grpc"
	"spider/internal/conf"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	configName string

	id, _ = os.Hostname()
)

func init() {
	configName = "config.yaml"
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf "+configName)
}

func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
		),
	)
}

func main() {
	flag.Parse()
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second,
		DialOptions: []google_grpc.DialOption{google_grpc.WithBlock()},
	})
	if err != nil {
		panic(err)
	}
	getwd, err := os.Getwd()
	_, fileName := filepath.Split(getwd)
	if err != nil {
		panic(err)
	}
	path := fileName + "-" + configName
	resp, err := client.Get(context.Background(), path, clientv3.WithLimit(1))
	if err != nil {
		panic(err)
	}
	if resp.Kvs == nil {
		f, err := os.Open(flagconf + "/" + configName)
		if err != nil {
			panic(err)
		}
		all, err := io.ReadAll(f)
		if err != nil {
			panic(err)
		}
		client.Put(context.Background(), path, string(all))
	}

	source, err := cfg.New(client, cfg.WithPath(path), cfg.WithPrefix(true))
	if err != nil {
		panic(err)
	}

	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	c := config.New(
		config.WithSource(
			source,
		),
		config.WithResolver(func(m map[string]interface{}) (err error) {
			return
		}),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

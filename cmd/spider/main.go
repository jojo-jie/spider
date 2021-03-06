package main

import (
	"flag"
	"fmt"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/sdk/trace"
	"os"
	"spider/internal/conf"
	"strings"
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
	pwd, _ := os.Getwd()
	Name = pwd[strings.LastIndex(pwd, "/")+1:]
}

func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server, provider *trace.TracerProvider, reg registry.Registrar) *kratos.App {
	options := make([]kratos.Option, 0, 10)
	options = append(options, kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
		))
	if reg != nil {
		options = append(options, kratos.Registrar(reg))
	}
	return kratos.New(
		options...,
	)
}

func main() {
	flag.Parse()

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
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger, Name)
	if err != nil {
		panic(fmt.Errorf("stack %+v", errors.WithStack(err)))
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(fmt.Errorf("stack %+v", errors.WithStack(err)))
	}
}

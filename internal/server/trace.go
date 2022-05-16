package server

import (
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"spider/internal/conf"
	"track"
)

// InitGlobalTracer set trace provider
func InitGlobalTracer(c *conf.Server, name string) (*tracesdk.TracerProvider, error) {
	return track.New(c.GetOpenTelemetry(), name)
}

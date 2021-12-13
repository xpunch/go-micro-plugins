package opentracing

import (
	"context"
	"fmt"

	opentracing "github.com/opentracing/opentracing-go"
	opentracinglog "github.com/opentracing/opentracing-go/log"
	"go-micro.dev/v4/client"
)

type publishWrapper struct {
	ot opentracing.Tracer
	client.Client
}

func (o *publishWrapper) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) error {
	name := fmt.Sprintf("Pub to %s", p.Topic())
	ctx, span, err := StartSpanFromContext(ctx, o.ot, name)
	if err != nil {
		return err
	}
	defer span.Finish()
	if err = o.Client.Publish(ctx, p, opts...); err != nil {
		span.LogFields(opentracinglog.String("error", err.Error()))
		span.SetTag("error", true)
	}
	return err
}

// NewPublishWrapper accepts an open tracing Trace and returns a Publish Wrapper
func NewPublishWrapper(ot opentracing.Tracer) client.Wrapper {
	return func(c client.Client) client.Client {
		if ot == nil {
			ot = opentracing.GlobalTracer()
		}
		return &publishWrapper{ot, c}
	}
}

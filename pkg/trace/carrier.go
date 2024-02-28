package trace

import (
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/metadata"
)

// propagator Carrier

var _ propagation.TextMapCarrier = (*Carrier)(nil)

type Carrier struct {
	MD *metadata.MD
}

func (c *Carrier) Get(key string) string {
	values := c.MD.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (c *Carrier) Set(key, value string) {
	c.MD.Set(key, value)
}

func (c *Carrier) Keys() []string {
	out := make([]string, 0, len(*c.MD))
	for key := range *c.MD {
		out = append(out, key)
	}
	return out
}

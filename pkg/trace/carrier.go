package trace

import "google.golang.org/grpc/metadata"

// tracer通过Inject和Extract将span context信息注入到carrier,以便在跨进程的span间传递,
// 跨进程传递自定义k/v通过Baggage实现

type Carrier struct {
	MD metadata.MD
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
	out := make([]string, 0, len(c.MD))
	for key := range c.MD {
		out = append(out, key)
	}
	return out
}

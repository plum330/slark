package trace

import "google.golang.org/grpc/metadata"

type Metadata struct {
	metadata *metadata.MD
}

func (m *Metadata) Get(key string) string {
	values := m.metadata.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (m *Metadata) Set(key, value string) {
	m.metadata.Set(key, value)
}

func (m *Metadata) Keys() []string {
	out := make([]string, 0, len(*m.metadata))
	for key := range *m.metadata {
		out = append(out, key)
	}
	return out
}

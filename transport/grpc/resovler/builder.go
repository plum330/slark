package resovler

import (
	utils "github.com/go-slark/slark/pkg"
	"google.golang.org/grpc/resolver"
)

type builder struct{}

func (b *builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	return nil, nil
}

func (b *builder) Scheme() string {
	return utils.Discovery
}

func NewBuilder() resolver.Builder {
	return nil
}

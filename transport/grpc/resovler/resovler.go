package resovler

import (
	"google.golang.org/grpc/resolver"
)

type resovler struct {
}

func (r *resovler) ResolveNow(opts resolver.ResolveNowOptions) {}

func (r *resovler) Close() {

}

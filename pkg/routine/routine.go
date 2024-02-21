package routine

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/logger"
	"sync"
)

func Go(ctx context.Context, fn func()) {
	defer func(ctx context.Context) {
		if r := recover(); r != nil {
			logger.Log(ctx, logger.ErrorLevel, map[string]interface{}{"error": fmt.Sprintf("%+v", r)}, "routine recover")
		}
	}(ctx)
	fn()
}

// multi routines composition

type Routine interface {
	Start()
}

type Group struct {
	routines []Routine
}

func NewGroup() *Group {
	return &Group{}
}

func (g *Group) Append(r ...Routine) {
	g.routines = append(g.routines, r...)
}

func (g *Group) Start() {
	wg := sync.WaitGroup{}
	wg.Add(len(g.routines))
	for index := range g.routines {
		go Go(context.TODO(), func() {
			g.routines[index].Start()
			wg.Done()
		})
	}
	wg.Wait()
}

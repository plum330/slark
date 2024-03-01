package routine

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/logger"
	"sync"
)

func GoSafe(ctx context.Context, fn func()) {
	go func() {
		defer func(ctx context.Context) {
			if r := recover(); r != nil {
				logger.Log(ctx, logger.ErrorLevel, map[string]interface{}{"error": fmt.Sprintf("%+v", r)}, "routine recover")
			}
		}(ctx)
		fn()
	}()
}

// multi routines composition

type Routine interface {
	Do()
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

func (g *Group) Do() {
	wg := sync.WaitGroup{}
	wg.Add(len(g.routines))
	for index := range g.routines {
		GoSafe(context.TODO(), func() {
			g.routines[index].Do()
			wg.Done()
		})
	}
	wg.Wait()
}

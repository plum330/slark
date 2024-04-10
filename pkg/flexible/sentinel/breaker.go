package sentinel

import (
	"encoding/json"
	"fmt"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/util"
	"os"
	"sync"
)

type Breaker struct {
	once  sync.Once
	mu    sync.RWMutex
	rules map[string]struct{}
}

type stateChangeTestListener struct{}

func (s *stateChangeTestListener) OnTransformToClosed(prev circuitbreaker.State, rule circuitbreaker.Rule) {
	fmt.Printf("rule.steategy: %+v, From %s to Closed, time: %d\n", rule.Strategy, prev.String(), util.CurrentTimeMillis())
}

func (s *stateChangeTestListener) OnTransformToOpen(prev circuitbreaker.State, rule circuitbreaker.Rule, snapshot interface{}) {
	fmt.Printf("rule.steategy: %+v, From %s to Open, snapshot: %d, time: %d\n", rule.Strategy, prev.String(), snapshot, util.CurrentTimeMillis())
}

func (s *stateChangeTestListener) OnTransformToHalfOpen(prev circuitbreaker.State, rule circuitbreaker.Rule) {
	fmt.Printf("rule.steategy: %+v, From %s to Half-Open, time: %d\n", rule.Strategy, prev.String(), util.CurrentTimeMillis())
}

func (b *Breaker) Init(path string) error {
	rules := make([]*circuitbreaker.Rule, 0)
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, &rules)
	if err != nil {
		return err
	}
	circuitbreaker.LoadRules(rules)
	b.mu.Lock()
	for _, rule := range rules {
		b.rules[rule.Resource] = struct{}{}
	}
	b.mu.Unlock()
	b.once.Do(func() {
		circuitbreaker.RegisterStateChangeListeners(&stateChangeTestListener{})
	})
	return nil
}

func (b *Breaker) Exists(resource string) bool {
	b.mu.RLock()
	_, exists := b.rules[resource]
	b.mu.RUnlock()
	return exists
}

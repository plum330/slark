package sentinel

import (
	"encoding/json"
	"github.com/alibaba/sentinel-golang/core/flow"
	"os"
	"sync"
)

type Flow struct {
	mu    sync.RWMutex
	rules map[string]struct{}
}

func (f *Flow) Init(path string) error {
	rules := make([]*flow.Rule, 0)
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, &rules)
	if err != nil {
		return err
	}
	flow.LoadRules(rules)
	f.mu.Lock()
	for _, rule := range rules {
		f.rules[rule.Resource] = struct{}{}
	}
	f.mu.Unlock()
	return nil
}

func (f *Flow) Exists(resource string) bool {
	f.mu.RLock()
	_, exists := f.rules[resource]
	f.mu.RUnlock()
	return exists
}

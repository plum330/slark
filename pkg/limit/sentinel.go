package limit

import (
	"encoding/json"
	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/flow"
	"os"
	"sync"
)

type Sentinel struct {
	l     sync.Mutex
	rules map[string]struct{}
}

func (s *Sentinel) init(path string) error {
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
	s.l.Lock()
	for _, rule := range rules {
		s.rules[rule.Resource] = struct{}{}
	}
	s.l.Unlock()
	return nil
}

func (s *Sentinel) exists(resource string) bool {
	s.l.Lock()
	_, exists := s.rules[resource]
	s.l.Unlock()
	return exists
}

func NewSentinel(path string) (*Sentinel, error) {
	s := &Sentinel{}
	if len(path) == 0 {
		return s, nil
	}
	err := s.init(path)
	if err != nil {
		return nil, err
	}
	entity := config.NewDefaultConfig()
	entity.Sentinel.App.Name = ""
	entity.Sentinel.Log.Dir = ""
	return s, api.InitWithConfig(entity)
}

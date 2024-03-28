package limit

import (
	"context"
	"encoding/json"
	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/fsnotify/fsnotify"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/pkg/routine"
	"os"
	"path/filepath"
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

func (s *Sentinel) watch(path string) {
	path, err := filepath.Abs(path)
	if err != nil {
		logger.Log(context.TODO(), logger.PanicLevel, map[string]interface{}{"error": err})
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Log(context.TODO(), logger.PanicLevel, map[string]interface{}{"error": err})
	}
	defer w.Close()
	f := filepath.Clean(path)
	links, _ := filepath.EvalSymlinks(path)

	routine.GoSafe(context.TODO(), func() {
		for {
			select {
			case event := <-w.Events:
				cLinks, _ := filepath.EvalSymlinks(path)
				logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{
					"event":     filepath.Clean(event.Name),
					"path":      filepath.Clean(path),
					"cur_links": cLinks,
					"links":     links,
				}, "watch event")
				// we only care about the config file with the following cases:
				// 1 - if the config file was modified or created
				// 2 - if the real path to the config file changed
				const writeOrCreateMask = fsnotify.Write | fsnotify.Create
				if (filepath.Clean(event.Name) == f && event.Op&writeOrCreateMask != 0) || (cLinks != "" && cLinks != links) {
					links = cLinks
					s.init(links)
					logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{"event": event.Name, "links": links}, "sentinel file modified")
				}
			case e := <-w.Errors:
				logger.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": e}, "watch error")
			}
		}
	})
	err = w.Add(path)
	if err != nil {
		logger.Log(context.TODO(), logger.PanicLevel, map[string]interface{}{"error": err}, "path error")
	}
	select {}
}

func (s *Sentinel) Exists(resource string) bool {
	s.l.Lock()
	_, exists := s.rules[resource]
	s.l.Unlock()
	return exists
}

// 全局sentinel

func NewSentinel(path string) (*Sentinel, error) {
	s := &Sentinel{}
	if len(path) != 0 {
		err := s.init(path)
		if err != nil {
			logger.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": err}, "init rule error")
		}
		go s.watch(path)
	}
	entity := config.NewDefaultConfig()
	entity.Sentinel.App.Name = "" // env
	return s, api.InitWithConfig(entity)
}

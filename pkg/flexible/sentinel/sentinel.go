package sentinel

import (
	"context"
	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/fsnotify/fsnotify"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/pkg/routine"
	"path/filepath"
	"sync"
)

type Resource interface {
	Init(path string) error
	Exists(resource string) bool
}

type Sentinel struct {
	Resource
}

type Option func(*Sentinel)

func New(path string, opts ...Option) (Resource, error) {
	s := &Sentinel{
		Resource: &Flow{
			mu:    sync.RWMutex{},
			rules: make(map[string]struct{}),
		},
	}
	for _, opt := range opts {
		opt(s)
	}
	if len(path) != 0 {
		err := s.Init(path)
		if err != nil {
			logger.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": err}, "init rule error")
		}
		go s.watch(path)
	}
	cfg := config.NewDefaultConfig()
	cfg.Sentinel.App.Name = "" // env
	return s, api.InitWithConfig(cfg)
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
	newPath, _ := filepath.EvalSymlinks(path)
	routine.GoSafe(context.TODO(), func() {
		for {
			select {
			case event := <-w.Events:
				curPath, _ := filepath.EvalSymlinks(path)
				logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{
					"event":    filepath.Clean(event.Name),
					"path":     filepath.Clean(path),
					"cur_path": curPath,
					"new_path": newPath,
				}, "watch event")
				// we only care about the config file with the following cases:
				// 1 - if the config file was modified or created
				// 2 - if the real path to the config file changed
				const writeOrCreateMask = fsnotify.Write | fsnotify.Create
				if (filepath.Clean(event.Name) == f && event.Op&writeOrCreateMask != 0) || (curPath != "" && curPath != newPath) {
					newPath = curPath
					s.Init(newPath)
					logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{"event": event.Name, "new_path": newPath}, "sentinel file modified")
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

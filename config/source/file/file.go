package file

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/pkg/routine"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type File struct {
	path   string
	dir    string
	notify chan struct{}
}

func NewFile(path string) *File {
	path, err := filepath.Abs(path)
	if err != nil {
		logger.Log(context.TODO(), logger.PanicLevel, map[string]interface{}{"error": err})
	}
	f := &File{
		path:   path,
		dir:    dir(path),
		notify: make(chan struct{}, 1),
	}
	routine.GoSafe(context.TODO(), func() {
		f.watch()
	})
	return f
}

func (f *File) watch() {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Log(context.TODO(), logger.PanicLevel, map[string]interface{}{"error": err})
	}
	defer w.Close()

	routine.GoSafe(context.TODO(), func() {
		for {
			select {
			case event := <-w.Events:
				logger.Log(context.TODO(), logger.InfoLevel, map[string]interface{}{
					"event": filepath.Clean(event.Name),
					"path":  filepath.Clean(f.path),
				})
				// we only care about the config file with the following cases:
				// 1 - if the config file was modified or created
				// 2 - if the real path to the config file changed
				const writeOrCreateMask = fsnotify.Write | fsnotify.Create
				if event.Op&writeOrCreateMask != 0 && filepath.Clean(event.Name) == filepath.Clean(f.path) {
					log.Println("modified file: ", event.Name)
					select {
					//case f.changed <- struct{}{}:
					default:
					}
				}
			case e := <-w.Errors:
				logger.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": e})
			}
		}
	})

	err = w.Add(f.dir)
	if err != nil {
		log.Fatal(err)
	}
	select {}
}

func (f *File) Load() ([]byte, error) {
	return os.ReadFile(f.path)
}

func (f *File) Watch() <-chan struct{} {
	return f.notify
}

func (f *File) Close() error {
	close(f.notify)
	return nil
}

func isDir(path string) (bool, error) {
	f, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	switch mode := f.Mode(); {
	case mode.IsDir():
		return true, nil
	case mode.IsRegular():
		return false, nil
	}
	return false, nil
}

func handleDir(dir string) string {
	if runtime.GOOS == "windows" {
		dir = strings.Replace(dir, "\\", "/", -1)
	}

	runes := []rune(dir)
	l := strings.LastIndex(dir, "/")
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[0:l])
}

func dir(path string) string {
	ok, err := isDir(path)
	if ok || err != nil {
		return path
	}
	return handleDir(path)
}

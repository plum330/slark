package pprof

import (
	_ "github.com/zeromicro/go-zero/core/proc"
	"mosn.io/holmes"
	"mosn.io/holmes/reporters/pyroscope_reporter"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"time"
)

// kill -usr1 pid

// kill -usr2 pid

func Init() {
	go http.ListenAndServe(":8082", nil)
}

type Holmes struct {
	endpoint string
	timeout  time.Duration
	name     string
}

type Option func(*Holmes)

func NewHolmes(opts ...Option) (*holmes.Holmes, error) {
	h := &Holmes{
		endpoint: "",
		timeout:  3 * time.Second,
		name:     "slark",
	}
	for _, opt := range opts {
		opt(h)
	}
	cfg := pyroscope_reporter.RemoteConfig{
		UpstreamAddress:        h.endpoint,
		UpstreamRequestTimeout: h.timeout,
	}
	hn, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	tags := map[string]string{"region": hn}
	reporter, err := pyroscope_reporter.NewPyroscopeReporter(h.name, tags, cfg, holmes.NewStdLogger())
	if err != nil {
		return nil, err
	}
	_, err = url.Parse(h.endpoint)
	if err != nil {
		reporter = nil
	}
	ho, err := holmes.New(
		holmes.WithProfileReporter(reporter),
		holmes.WithCollectInterval("5s"),
		holmes.WithDumpPath("/tmp"),
		holmes.WithTextDump(),
		holmes.WithMemoryLimit(100*1024*1024), // 100MB
		holmes.WithCPUMax(85),
		// profile
		holmes.WithCPUDump(20, 100, 150, time.Minute*2),
		holmes.WithMemDump(50, 100, 800, time.Minute),
		holmes.WithGoroutineDump(200, 100, 5000, 200*5000, time.Minute),
		holmes.WithGCHeapDump(10, 20, 40, time.Minute),
	)
	ho.EnableCPUDump().EnableGoroutineDump().EnableMemDump().EnableGCHeapDump().Start()
	return ho, nil
}

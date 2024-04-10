package profiling

import (
	"mosn.io/holmes"
	"mosn.io/holmes/reporters/pyroscope_reporter"
	"net/url"
	"os"
	"time"
)

type Dump struct {
	min      int
	diff     int
	abs      int
	max      int
	coolDown time.Duration
	enable   bool
}

type Report struct {
	endpoint string
	timeout  time.Duration
	name     string
}

type Holmes struct {
	interval          string
	path              string
	memLimit          uint64
	cpuMax            int
	report            Report
	cpu, mem, cor, gc Dump
}

type Option func(*Holmes)

func WithInterval(interval string) Option {
	return func(h *Holmes) {
		h.interval = interval
	}
}

func WithPath(path string) Option {
	return func(h *Holmes) {
		h.path = path
	}
}

func WithMemLimit(limit uint64) Option {
	return func(h *Holmes) {
		h.memLimit = limit
	}
}

func WithCpuMax(max int) Option {
	return func(h *Holmes) {
		h.cpuMax = max
	}
}

func WithReport(report Report) Option {
	return func(h *Holmes) {
		h.report = report
	}
}

func WithCpu(cpu Dump) Option {
	return func(h *Holmes) {
		h.cpu = cpu
	}
}

func WithMem(mem Dump) Option {
	return func(h *Holmes) {
		h.mem = mem
	}
}

func WithCor(cor Dump) Option {
	return func(h *Holmes) {
		h.cor = cor
	}
}

func WithGC(gc Dump) Option {
	return func(h *Holmes) {
		h.gc = gc
	}
}

func NewHolmes(opts ...Option) (*holmes.Holmes, error) {
	h := &Holmes{
		interval: "5s",
		path:     "/tmp",
		memLimit: 100 * 1024 * 1024, // 100MB -> 1%
		cpuMax:   85,
		report: Report{
			endpoint: "",
			timeout:  3 * time.Second,
			name:     "slark",
		},
		cpu: Dump{
			min:      20,
			diff:     25,
			abs:      80,
			coolDown: time.Minute,
			enable:   true,
		},
		cor: Dump{
			min:      10,
			diff:     25,
			abs:      2000,
			max:      100 * 1000,
			coolDown: time.Minute,
			enable:   true,
		},
		mem: Dump{
			min:      30,
			diff:     25,
			abs:      80,
			coolDown: time.Minute,
			enable:   true,
		},
		gc: Dump{
			min:      10,
			diff:     20,
			abs:      40,
			coolDown: time.Minute,
			enable:   false,
		},
	}
	for _, opt := range opts {
		opt(h)
	}
	cfg := pyroscope_reporter.RemoteConfig{
		UpstreamAddress:        h.report.endpoint,
		UpstreamRequestTimeout: h.report.timeout,
	}
	hn, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	tags := map[string]string{"region": hn}
	reporter, err := pyroscope_reporter.NewPyroscopeReporter(h.report.name, tags, cfg, holmes.NewStdLogger())
	if err != nil {
		return nil, err
	}
	_, err = url.Parse(h.report.endpoint)
	if err != nil {
		reporter = nil
	}
	ho, err := holmes.New(
		holmes.WithProfileReporter(reporter),
		holmes.WithCollectInterval(h.interval),
		holmes.WithDumpPath(h.path),
		holmes.WithTextDump(),
		holmes.WithMemoryLimit(h.memLimit),
		holmes.WithCPUMax(h.cpuMax),
		//holmes.WithCPUCore(2),

		// profile
		// cup_usage > min% && cup_usage > (100 + diff)% * previous cpu usage recorded || cup_usage > abs%
		holmes.WithCPUDump(h.cpu.min, h.cpu.diff, h.cpu.abs, h.cpu.coolDown),
		// min < cur_g < max && cur_g > (100 + diff)% * previous_average_goroutine_num || cur_g > abs
		holmes.WithGoroutineDump(h.cor.min, h.cor.diff, h.cor.abs, h.cor.max, h.cor.coolDown),
		// mem_usage > min% && mem_usage > (100+diff)% * previous memory usage || mem_usage > abs%
		holmes.WithMemDump(h.mem.min, h.mem.max, h.mem.abs, h.mem.coolDown),
		holmes.WithGCHeapDump(h.gc.min, h.gc.diff, h.gc.abs, h.gc.coolDown),
	)
	if h.cpu.enable {
		ho = ho.EnableCPUDump()
	}
	if h.cor.enable {
		ho = ho.EnableGoroutineDump()
	}
	if h.mem.enable {
		ho = ho.EnableMemDump()
	}
	if h.gc.enable {
		ho = ho.EnableGCHeapDump()
	}
	ho.Start()
	return ho, nil
}

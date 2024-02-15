package balancer

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const Balancer = "balancer"

var pickerBuilder = &builder{Builder: NewRandomBuilder()}

func SetBuilder(builder Builder) {
	pickerBuilder.Builder = builder
}

func init() {
	balancer.Register(base.NewBalancerBuilder(
		Balancer,
		pickerBuilder,
		base.Config{HealthCheck: true},
	))
}

type Builder interface {
	Build() Picker
}

type builder struct {
	Builder
}

func (b *builder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	p := &picker{
		Picker: b.Builder.Build(),
	}
	p.Update(info.ReadySCs)
	return p
}

type Picker interface {
	Update(map[balancer.SubConn]base.SubConnInfo)
	Pick() balancer.SubConn
}

type picker struct {
	Picker
}

func (p *picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	pr := balancer.PickResult{
		SubConn: p.Picker.Pick(),
		Done: func(info balancer.DoneInfo) {
			// TODO
		},
	}
	return pr, nil
}

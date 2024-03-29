package apollo

import (
	"github.com/go-slark/slark/encoding/properties"
	"github.com/philchia/agollo/v4"
)

type Apollo struct {
	client    agollo.Client
	namespace string
	notify    chan struct{}
}

func New(c *agollo.Conf) *Apollo {
	client := agollo.NewClient(c)
	client.Start()
	ap := &Apollo{
		client:    client,
		namespace: c.NameSpaceNames[0], // TODO
		notify:    make(chan struct{}, 1),
	}
	client.OnUpdate(func(e *agollo.ChangeEvent) {
		ap.notify <- struct{}{}
	})
	return ap
}

func (a *Apollo) Load() ([]byte, error) {
	cfg := a.client.GetContent(agollo.WithNamespace(a.namespace))
	return []byte(cfg), nil
}

func (a *Apollo) Watch() <-chan struct{} {
	return a.notify
}

func (a *Apollo) Close() error {
	return a.client.Stop()
}

func (a *Apollo) Format() string {
	return properties.Name
}

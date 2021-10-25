package component

import (
	"os"

	"github.com/logsquaredn/geocloud"
	"github.com/tedsuo/ifrit"
)

func NewNamedComponentFunc(name string, fn ifrit.RunFunc) geocloud.Component {
	return &componentFunc{
		fn:   fn,
		name: name,
	}
}

func NewComponentFunc(fn ifrit.RunFunc) geocloud.Component {
	return NewNamedComponentFunc("__func", fn) // do not group an unnamed func
}

type componentFunc struct {
	fn   ifrit.RunFunc
	name string
}

var _ geocloud.Component = (*componentFunc)(nil)

func (c *componentFunc) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	return c.fn.Run(signals, ready)
}

func (c *componentFunc) Execute(_ []string) error {
	return <-ifrit.Invoke(c).Wait()
}

func (c *componentFunc) Name() string {
	return c.name
}

func (c *componentFunc) IsConfigured() bool {
	return true
}

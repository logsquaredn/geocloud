package component

import (
	"os"

	"github.com/logsquaredn/geocloud"
	"github.com/tedsuo/ifrit"
)

// NewNamedComponentFunc returns a new component with the given name
// from the given ifrit.RunFunc
func NewNamedComponentFunc(name string, fn ifrit.RunFunc) geocloud.Component {
	return &componentFunc{
		fn:   fn,
		name: name,
	}
}

// NewComponentFunc returns a new component
// from the given ifrit.RunFunc
//
// If you intend to group a component returned from this function
// further, please use NewNamedComponentFunc
func NewComponentFunc(fn ifrit.RunFunc) geocloud.Component {
	return NewNamedComponentFunc("__func", fn) // do not group an unnamed func
}

type componentFunc struct {
	fn   ifrit.RunFunc
	name string
}

var _ geocloud.Component = (*componentFunc)(nil)

// Run executes the component's ifrit.RunFunc
func (c *componentFunc) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	return c.fn.Run(signals, ready)
}

// Execute runs the component and makes it implement the
// flags.Commander interface which lets it be a subcommand
func (c *componentFunc) Execute(_ []string) error {
	return <-ifrit.Invoke(c).Wait()
}

// Name returns the name of the component to be used to
// identify it when running as a part of a group of components
func (c *componentFunc) Name() string {
	return c.name
}

// IsEnabled reports true; this packaged currently
// provides no way to tell if a component was enabled
func (c *componentFunc) IsEnabled() bool {
	return true
}

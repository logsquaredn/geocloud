package component

import (
	"os"

	"github.com/logsquaredn/geocloud"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
)

type group struct {
	cs   []geocloud.Component
	name string
}

var _ geocloud.Component = (*group)(nil)

// Run executes the Run function of each of the group's components in order
func (g *group) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	ms := make([]grouper.Member, len(g.cs))
	for i := range ms {
		c := g.cs[i]
		ms[i] = grouper.Member{
			Name:   c.Name(),
			Runner: c,
		}
	}

	return grouper.NewOrdered(os.Interrupt, ms).Run(signals, ready)
}

// Execute runs the component and makes it implement the
// flags.Commander interface which lets it be a subcommand
func (g *group) Execute(_ []string) error {
	return <-ifrit.Invoke(g).Wait()
}

// Name returns the name of the component to be used to
// identify it when running as a part of a group of components
func (g *group) Name() string {
	return g.name
}

// IsEnabled reports whether or not the component is
// fully configured to be ran
func (g *group) IsEnabled() bool {
	for _, c := range g.cs {
		if !c.IsEnabled() {
			return false
		}
	}
	return true
}

// NewNamedGroup returns a geocloud.Component that is a group
// of the components passed to it with the given name
func NewNamedGroup(name string, cs ...geocloud.Component) geocloud.Component {
	return &group{cs: cs, name: name}
}

// NewGroup returns a geocloud.Component that is a group
// of the components passed to it
func NewGroup(cs ...geocloud.Component) geocloud.Component {
	return NewNamedGroup("__group", cs...) // do not group an unnamed group
}

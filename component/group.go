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

func (g *group) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	ms := make([]grouper.Member, len(g.cs))
	for i := range ms {
		c := g.cs[i]
		ms[i] = grouper.Member{
			Name: c.Name(),
			Runner: c,
		}
	}

	return grouper.NewOrdered(os.Interrupt, ms).Run(signals, ready)
}

func (g *group) Execute(_ []string) error {
	return <-ifrit.Invoke(g).Wait()
}

func (g *group) Name() string {
	return g.name
}

func (g *group) IsConfigured() bool {
	for _, c := range g.cs {
		if !c.IsConfigured() {
			return false
		}
	}
	return true
}


// NamedGroup returns a geocloud.Component that is a group
// of the components passed to it with the given name
func NewNamedGroup(name string, cs ...geocloud.Component) geocloud.Component {
	return &group{ cs: cs, name: name }
}

// Group returns a geocloud.Component that is a group
// of the components passed to it
func NewGroup(cs ...geocloud.Component) geocloud.Component {
	return NewNamedGroup("__group", cs...) // do not group an unnamed group
}
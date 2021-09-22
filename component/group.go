package component

import (
	"os"

	"github.com/logsquaredn/geocloud"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
)

type group struct {
	cs []geocloud.Component	
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

	return <-ifrit.Invoke(grouper.NewOrdered(os.Interrupt, ms)).Wait()
}

func (g *group) Execute(_ []string) error {
	return <-ifrit.Invoke(g).Wait()
}

func (g *group) Name() string {
	return "_group" // do not Group a group
}

func (g *group) IsConfigured() bool {
	for _, c := range g.cs {
		if !c.IsConfigured() {
			return false
		}
	}
	return true
}

// Group returns a geocloud.Component that is a group
// of the components passed to it
func Group(cs ...geocloud.Component) geocloud.Component {
	return &group{ cs: cs }
}

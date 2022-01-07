package component

import "github.com/logsquaredn/geocloud"

func Coalesce(cs ...geocloud.Component) geocloud.Component {
	for _, c := range cs {
		if c.IsEnabled() {
			return c
		}
	}

	return nil
}

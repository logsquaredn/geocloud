package api

import "github.com/logsquaredn/geocloud"

type message struct {
	id string
}

var _ geocloud.Message = (*message)(nil)

func (m *message) ID() string {
	return m.id
}

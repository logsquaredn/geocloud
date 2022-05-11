package messagequeue

import "github.com/logsquaredn/geocloud"

type message struct {
	id string
}

var _ geocloud.Message = (*message)(nil)

func (m *message) GetID() string {
	return m.id
}

package geocloud

// Message is the primary object that is passed between geocloud components
type Message interface {
	GetID() string
}

type Msg string

func (m Msg) GetID() string {
	return string(m)
}

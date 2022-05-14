package geocloud

// Message is the primary object that is passed between geocloud components
type Message interface {
	GetID() string
}

func NewMessage(id string) Message {
	return message(id)
}

type message string

func (m message) GetID() string {
	return string(m)
}

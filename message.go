package rototiller

// Message is the primary object that is passed between rototiller components
type Message interface {
	GetID() string
}

type Msg string

func (m Msg) GetID() string {
	return string(m)
}

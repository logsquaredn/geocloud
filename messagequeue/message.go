package messagequeue

type message string

func (m message) GetID() string {
	return string(m)
}

package geocloud

type Message interface {
	ID() string
}

type Job struct {
	ID       string
	TaskType string
	Status   string
	Error    error
}

type Task struct {
	Type      string
	Params    []string
	QueueName string
	Ref       string
}

const (
	Completed  = "COMPLETED"
	InProgress = "IN PROGRES"
	Waiting    = "WAITING"
	Error      = "ERROR"
)

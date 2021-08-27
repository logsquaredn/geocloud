package geocloud

type Message interface {
	ID() string
}

type Job struct {
	ID        string
	TaskType  string
	Status    string
	Error     error
	StartTime string
	EndTime   string
	Params    string
}

type Task struct {
	Type      string
	Params    []string
	QueueName string
	Ref       string
}

const (
	Completed  = "COMPLETED"
	InProgress = "IN PROGRESS"
	Waiting    = "WAITING"
	Error      = "ERROR"
)

package api

type EventType string

func (e EventType) String() string {
	return string(e)
}

const (
	EventTypeJobCreated   EventType = "job.created"
	EventTypeJobStarted   EventType = "job.started"
	EventTypeJobCompleted EventType = "job.completed"
	EventTypeJobAny       EventType = "job.#"

	EventTypeStorageCreated EventType = "storage.created"
	EventTypeStorageAny     EventType = "storage.#"

	EventTypeAny EventType = "#"
)

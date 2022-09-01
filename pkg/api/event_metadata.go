package api

type JobEventMetadata map[string]string

func (m JobEventMetadata) GetId() string {
	return m["id"]
}

package api

import (
	"fmt"
	"strings"
)

type JobStatus string

const (
	JobStatusWaiting    JobStatus = "waiting"
	JobStatusInProgress JobStatus = "inprogress"
	JobStatusComplete   JobStatus = "complete"
	JobStatusError      JobStatus = "error"
)

func (s JobStatus) String() string {
	return string(s)
}

func ParseJobStatus(jobStatus string) (JobStatus, error) {
	for _, j := range []JobStatus{
		JobStatusWaiting, JobStatusInProgress,
		JobStatusComplete, JobStatusError,
	} {
		if strings.EqualFold(jobStatus, j.String()) {
			return j, nil
		}
	}
	return "", fmt.Errorf("unknown job status '%s'", jobStatus)
}

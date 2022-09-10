package rototiller

import "github.com/logsquaredn/rototiller/pkg/api"

const (
	StorageStatusUnknown       = api.StorageStatusUnknown
	StorageStatusFinal         = api.StorageStatusFinal
	StorageStatusUnusable      = api.StorageStatusUnusable
	StorageStatusTransformable = api.StorageStatusTransformable
)

const (
	JobStatusWaiting    = api.JobStatusWaiting
	JobStatusInProgress = api.JobStatusInProgress
	JobStatusComplete   = api.JobStatusComplete
	JobStatusError      = api.JobStatusError
)

const (
	TaskKindLookup = api.TaskKindLookup
)

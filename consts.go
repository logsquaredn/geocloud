package rototiller

import "github.com/logsquaredn/rototiller/pb"

const (
	StorageStatusUnknown       = pb.StorageStatusUnknown
	StorageStatusFinal         = pb.StorageStatusFinal
	StorageStatusUnusable      = pb.StorageStatusUnusable
	StorageStatusTransformable = pb.StorageStatusTransformable
)

const (
	JobStatusWaiting    = pb.JobStatusWaiting
	JobStatusInProgress = pb.JobStatusInProgress
	JobStatusComplete   = pb.JobStatusComplete
	JobStatusError      = pb.JobStatusError
)

const (
	TaskKindLookup = pb.TaskKindLookup
)

package api

import (
	"fmt"
	"strings"
)

type TaskType string

const (
	TaskTypeBuffer            TaskType = "buffer"
	TaskTypeFilter            TaskType = "filter"
	TaskTypeRemoveBadGeometry TaskType = "removebadgeometry"
	TaskTypeReproject         TaskType = "reproject"
	TaskTypeVectorLookup      TaskType = "vectorlookup"
	TaskTypeRasterLookup      TaskType = "rasterlookup"
)

var (
	AllTaskTypes = []TaskType{
		TaskTypeBuffer, TaskTypeFilter, TaskTypeRemoveBadGeometry,
		TaskTypeReproject, TaskTypeVectorLookup, TaskTypeRasterLookup,
	}
)

func (t TaskType) String() string {
	return string(t)
}

func ParseTaskType(taskType string) (TaskType, error) {
	for _, t := range AllTaskTypes {
		if strings.EqualFold(taskType, t.String()) {
			return t, nil
		}
	}

	return "", fmt.Errorf("unknown task type '%s'", taskType)
}

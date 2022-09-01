package api

import (
	"fmt"
	"strings"
)

type StorageStatus string

const (
	StorageStatusFinal         StorageStatus = "final"
	StorageStatusUnknown       StorageStatus = "unknown"
	StorageStatusUnusable      StorageStatus = "unusable"
	StorageStatusTransformable StorageStatus = "transformable"
)

func (k StorageStatus) String() string {
	return string(k)
}

func ParseStorageStatus(storageStatus string) (StorageStatus, error) {
	for _, k := range []StorageStatus{
		StorageStatusFinal, StorageStatusUnknown,
		StorageStatusUnusable, StorageStatusTransformable,
	} {
		if strings.EqualFold(storageStatus, k.String()) {
			return k, nil
		}
	}

	return "", fmt.Errorf("unknown storage status '%s'", storageStatus)
}

package pb

import (
	"fmt"
	"strings"
)

type TaskKind string

const (
	TaskKindLookup         TaskKind = "lookup"
	TaskKindTransformation TaskKind = "transformation"
)

func (k TaskKind) String() string {
	return string(k)
}

func ParseTaskKind(taskKind string) (TaskKind, error) {
	for _, k := range []TaskKind{TaskKindLookup, TaskKindTransformation} {
		if strings.EqualFold(taskKind, k.String()) {
			return k, nil
		}
	}

	return "", fmt.Errorf("unknown task kind '%s'", taskKind)
}

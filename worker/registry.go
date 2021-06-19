package worker

import (
	"fmt"
	"strings"
)

type Registry string

func (r Registry) Ref(image string) string {
	if strings.Contains(image, "/") {
		if strings.Contains(image, "@") || strings.Contains(image, ":") {
			return fmt.Sprintf("%s/%s", r, image)
		}
		return fmt.Sprintf("%s/%s:latest", r, image)
	} else if strings.Contains(image, "@") || strings.Contains(image, ":") {
		return fmt.Sprintf("%s/library/%s", r, image)
	}
	return fmt.Sprintf("%s/library/%s:latest", r, image)
}

func NewRegistry(registry string) (*Registry, error) {
	ret := Registry(strings.Trim(registry, " /"))
	return &ret, nil
}

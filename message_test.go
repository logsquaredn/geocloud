package geocloud_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/logsquaredn/geocloud"
	"github.com/stretchr/testify/assert"
)

func TestMsg(t *testing.T) {
	var (
		expected = uuid.NewString()
		actual   = geocloud.Msg(expected)
	)
	assert.Equal(t, expected, actual.GetID())
}

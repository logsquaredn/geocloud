package geocloud_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/logsquaredn/geocloud"
	"github.com/stretchr/testify/assert"
)


func TestNewMessage(t *testing.T) {
	var (
		expected = uuid.NewString()
		actual   = geocloud.NewMessage(expected)
	)
	assert.Equal(t, expected, actual.GetID())
}
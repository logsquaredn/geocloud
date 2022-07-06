package rototiller_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/logsquaredn/rototiller"
	"github.com/stretchr/testify/assert"
)

func TestMsg(t *testing.T) {
	var (
		expected = uuid.NewString()
		actual   = rototiller.Msg(expected)
	)
	assert.Equal(t, expected, actual.GetID())
}

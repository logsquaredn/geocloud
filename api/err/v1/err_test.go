package errv1_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/logsquaredn/geocloud/api/err/v1"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	var (
		expected1, expected2, expected3 = errors.New(uuid.NewString()), 1, 2
		actual                          = errv1.New(expected1, expected2, expected3)
	)
	assert.Equal(t, expected1.Error(), actual.Error())
	assert.Equal(t, expected2, actual.HTTPStatusCode)
	assert.Equal(t, expected3, int(actual.ConnectCode))
}

func TestErrorWithError(t *testing.T) {
	var (
		expected1 = errv1.New(errors.New("test"), 1, 2)
		actual    = errv1.New(expected1)
	)
	assert.Equal(t, expected1, actual)
}

func TestErrorWithNil(t *testing.T) {
	var (
		actual = errv1.New(nil, 1, 2)
	)
	assert.Nil(t, actual)
	assert.Equal(t, "", actual.Error())
}

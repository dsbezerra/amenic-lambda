package jobs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckOpeningMovies(t *testing.T) {
	data, err := mockDataAccessLayer()
	assert.NoError(t, err)
	assert.NotNil(t, data)
	err = CheckOpeningMovies(nil, data)
	assert.NoError(t, err)
}

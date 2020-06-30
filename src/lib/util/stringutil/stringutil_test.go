package stringutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToLowerTrimmed(t *testing.T) {
	expected := "aves de rapina"
	assert.Equal(t, expected, ToLowerTrimmed("Aves de Rapina"))
	assert.Equal(t, expected, ToLowerTrimmed("  Aves de Rapina"))
	assert.Equal(t, expected, ToLowerTrimmed("Aves de Rapina  "))
	assert.Equal(t, expected, ToLowerTrimmed("  Aves de Rapina  "))
	assert.Equal(t, "", ToLowerTrimmed(" "))
}

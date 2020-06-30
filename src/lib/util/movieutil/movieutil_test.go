package movieutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSlug(t *testing.T) {
	input := "Midway - Batalha em Alto Mar"
	expected := "midway-batalha-em-alto-mar"
	actual := GenerateSlug(input, true)
	assert.Equal(t, expected, actual)
}

func TestCapTitle(t *testing.T) {
	input := "Playmobil - O Filme"
	expected := "Playmobil - O Filme"
	actual := CapTitle(input)
	assert.Equal(t, expected, actual)
}

func TestUnromanTitle(t *testing.T) {
	input := "Frozen II"
	expected := "Frozen 2"
	actual := UnromanTitle(input)
	assert.Equal(t, expected, actual)
}

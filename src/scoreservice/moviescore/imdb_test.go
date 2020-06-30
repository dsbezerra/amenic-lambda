package moviescore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImdbSearch(t *testing.T) {
	query := "iron man 2008"

	imdb := NewIMDb()
	result, err := imdb.Search(query)
	assert.NoError(t, err)

	expectedFirstItem := ResultItem{
		ID:    "tt0371746",
		Title: "Iron Man",
		Year:  2008,
	}
	assert.NotEmpty(t, result.Items)
	assert.Equal(t, expectedFirstItem.ID, result.Items[0].ID)
}

func TestImdbScore(t *testing.T) {
	ID := "tt0371746"

	imdb := NewIMDb()
	result, err := imdb.Score(ID)
	assert.NoError(t, err)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, float32(7.9), result.Items[0].Score)
}

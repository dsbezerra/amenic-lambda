package jobs

import (
	"testing"

	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/mongolayer"

	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/stretchr/testify/assert"
)

const (
	testConnection = "mongodb+srv://admin:HPgNbJ3aZAWVTAHJ@cluster0-2uphl.mongodb.net/amenic-dev?retryWrites=true&w=majority"
)

func TestParseArgs(t *testing.T) {
	fakeJobInput := &Input{
		Name: "create_static",
		Args: []string{"-type", "now_playing,schedule,upcoming"},
	}

	args := parseArgs(fakeJobInput)
	assert.NotNil(t, args)
	assert.NotNil(t, args["type"])
	assert.Equal(t, fakeJobInput.Args[1], args["type"])
}

func mockDataAccessLayer() (persistence.DataAccessLayer, error) {
	data, err := mongolayer.NewMongoDAL(testConnection)
	if err != nil {
		return nil, err
	}
	return data, nil
}

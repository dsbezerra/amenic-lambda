package jobs

import (
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/dsbezerra/amenic-lambda/src/scoreservice/task"
)

// SyncScores ...
func SyncScores(input *Input, data persistence.DataAccessLayer) error {
	return task.SyncScores(data)
}

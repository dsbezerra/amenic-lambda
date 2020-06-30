package jobs

// CheckOpeningMovies ...
import (
	"time"

	"github.com/dsbezerra/amenic-lambda/src/lib/util/scheduleutil"

	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/models"
	"github.com/dsbezerra/amenic-lambda/src/notificationservice/fcm"
	"github.com/dsbezerra/amenic-lambda/src/notificationservice/weekreleases"
	"github.com/pkg/errors"
)

var (
	// ErrNoMoviesPlaying ...
	ErrNoMoviesPlaying = errors.New("there are no movies currently playing")
)

// CheckOpeningMovies simply checks for releases and sends a notification.
func CheckOpeningMovies(in *Input, data persistence.DataAccessLayer) error {
	releases, err := GetWeekReleases(data)
	if err != nil {
		return err
	}
	notification := weekreleases.PrepareNotification(releases)
	if notification != nil {
		_, err := fcm.SendNotification(notification)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetWeekReleases ...
func GetWeekReleases(data persistence.DataAccessLayer) ([]models.Movie, error) {
	result := make([]models.Movie, 0)

	playing, err := data.GetNowPlayingMovies(data.DefaultQuery().AddInclude("theaters"))
	if err != nil {
		return nil, err
	}

	if len(playing) == 0 {
		return nil, ErrNoMoviesPlaying
	}

	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	week := scheduleutil.GetWeekPeriod(&now)
	start := week.Start.UnixNano()
	end := week.End.UnixNano()
	for _, m := range playing {
		if m.Hidden || m.ReleaseDate == nil {
			continue
		}

		release := m.ReleaseDate.UnixNano()
		if release >= start && release <= end {
			result = append(result, m)
		}
	}

	return result, nil
}

package extractors

import (
	"log"
	"time"

	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/models"
	"github.com/dsbezerra/amenic-lambda/src/lib/util/scheduleutil"
	"github.com/dsbezerra/amenic-lambda/src/lib/util/scraperutil"
	"github.com/dsbezerra/amenic-lambda/src/lib/util/timeutil"
	"github.com/dsbezerra/amenic-lambda/src/scraperservice/provider"
	"go.mongodb.org/mongo-driver/bson"
)

type (
	ScheduleExtractor struct {
		Data     persistence.DataAccessLayer
		Provider provider.Provider
		Run      *models.ScraperRun
		Sessions []models.Session
	}
)

// NewScheduleExtractor ...
func NewScheduleExtractor(data persistence.DataAccessLayer, p provider.Provider, s *models.ScraperRun) *ScheduleExtractor {
	result := &ScheduleExtractor{
		Data:     data,
		Run:      s,
		Provider: p,
	}
	return result
}

// Execute extract schedule information for a given provider
func (e *ScheduleExtractor) Execute() error {
	result, err := e.Provider.GetSchedule()
	if err != nil {
		return err
	}

	movies := LookupMovies(e.Data, result)
	m := map[string]models.Movie{}
	for _, movie := range movies {
		m[movie.Slugs.NoDashes] = movie
	}

	for i := range result {
		v, ok := m[result[i].MovieSlugs.NoDashes]
		if ok {
			result[i].MovieID = v.ID
			result[i].MovieSlugs = v.Slugs
		}
		result[i].Movie = nil
	}
	e.Sessions = result
	return nil
}

// Complete TODO
func (e *ScheduleExtractor) Complete() {

	switch e.Run.ResultCode {
	case scraperutil.RunResultSuccess:
		now := time.Now()
		start := scheduleutil.GetWeekPeriod(&now).Start
		query := e.Data.DefaultQuery().AddCondition("$and", []bson.D{
			bson.D{
				{Key: "theaterId", Value: e.Run.Scraper.TheaterID},
				{Key: "startTime", Value: bson.D{
					{Key: "$gte", Value: start},
				}},
			},
		})
		_, err := e.Data.DeleteSessions(query)
		if err != nil {
			// TODO: Handle
			log.Fatal(err)
		}
		err = e.Data.InsertSessions(e.Sessions...)
		if err != nil {
			// TODO: Handle
			log.Fatal(err)
		}
	case scraperutil.RunResultNotModified:
		fallthrough
	default:
		// Do nothing.
	}
}

// ExtractedHash TODO
func (e *ScheduleExtractor) ExtractedHash() string {
	return GetExtractedHash(e.Sessions)
}

// ExtractedCount TODO
func (e *ScheduleExtractor) ExtractedCount() int {
	return len(e.Sessions)
}

// LookupMovies ...
type PairSlugMovie struct {
	Slugs models.Slugs
	Movie models.Movie
}

func LookupMovies(data persistence.DataAccessLayer, sessions []models.Session) []models.Movie {
	defer timeutil.TimeTrack(time.Now(), "LookupMovies")

	claquete := []int{}
	slugs := []string{}

	movies := make([]models.Movie, 0)

	seen := map[string]PairSlugMovie{}
	for _, s := range sessions {
		if s.MovieSlugs.NoDashes != "" {
			_, ok := seen[s.MovieSlugs.NoDashes]
			if !ok {
				seen[s.MovieSlugs.NoDashes] = PairSlugMovie{
					Slugs: s.MovieSlugs,
					Movie: *s.Movie,
				}
			}
		}
	}

	for _, p := range seen {
		m := p.Movie
		if m.ClaqueteID != 0 {
			claquete = append(claquete, m.ClaqueteID)
		} else {
			slugs = append(slugs, p.Slugs.NoDashes)
		}
	}

	if len(claquete) > 0 {
		// @Refactor support other databases.
		m, err := data.GetMovies(data.DefaultQuery().AddCondition("claqueteId", bson.M{"$in": claquete}))
		if err == nil {
			movies = append(movies, m...)
		}
	}

	if len(slugs) > 0 {
		// @Refactor support other databases.
		m, err := data.GetMovies(data.DefaultQuery().AddCondition("slugs.noDashes", bson.M{"$in": slugs}))
		if err == nil {
			movies = append(movies, m...)
		}
	}

	// @Improve this check
	if len(movies) == len(seen) {
		return movies
	}

	// We still need to find movies
	f := map[string]models.Movie{}
	for _, m := range movies {
		f[m.Slugs.NoDashes] = m
	}

	for k, p := range seen {
		_, ok := f[k]
		if !ok {
			found, movie := FindMovieMatch(data, &p.Movie)
			if found {
				movies = append(movies, *movie)
			}
		}
	}

	return movies
}

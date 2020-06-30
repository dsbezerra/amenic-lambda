package jobs

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dsbezerra/amenic-lambda/src/lib/env"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/models"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/mongolayer"
	"github.com/dsbezerra/amenic-lambda/src/lib/util/scheduleutil"
	"github.com/dsbezerra/amenic-lambda/src/lib/util/timeutil"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TriggerCreateStatic sends a message to our AmenicJobQueue
// to execute a create_static job with type 'home'.
// It should execute create_static for now_playing and upcoming too.
// func TriggerCreateStatic() error {
// 	attrs := map[string]*sqs.MessageAttributeValue{
// 		"Job": &sqs.MessageAttributeValue{
// 			DataType:    aws.String("String"),
// 			StringValue: aws.String(JobCreateStatic),
// 		},
// 		"Type": &sqs.MessageAttributeValue{
// 			DataType:    aws.String("String"),
// 			StringValue: aws.String("home"),
// 		},
// 	}
// 	messageId, err := SendMessageToQueue(&sqs.SendMessageInput{
// 		MessageAttributes: attrs,
// 		MessageBody:       aws.String("Information about which create_static operation to run."),
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("Success", messageId)
// 	return nil
// }

//
// Copy 'n paste from apiservice/v1/static.go
//

// StaticType enum-like type to represent each static file.
type StaticType uint

const (
	// StaticTypeHome refers to home.json static file.
	StaticTypeHome StaticType = iota
	// StaticTypeNowPlaying refers to now_playing.json static file.
	StaticTypeNowPlaying
	// StaticTypeUpcoming refers to upcoming.json static file.
	StaticTypeUpcoming
	// StaticTypeSize is used to know the size of these enum-like type.
	StaticTypeSize
)

// SessionType represents the type of sessions some theaters may
// be playing for a now playing movie in the current date.
//
// For example: preview, premiere or normal.
//
// normal   - an ordinary movie session.
// premiere - movie's sessions in the same day as release.
// preview  - movie's sessions before its release.
//
type SessionType uint

const (
	// SessionTypeNone indicates the now playing movie has no session type.
	SessionTypeNone SessionType = iota
	// SessionTypeNormal indicates the session is an ordinary one.
	SessionTypeNormal
	// SessionTypePremiere indicates the session are the very first one.
	// This equals to ESTREIA em pt-BR.
	SessionTypePremiere
	// SessionTypePreview indicates the movie will screen before its release date.
	// This equals to PRÉ-ESTREIA in pt-BR.
	SessionTypePreview
)

type (
	// StaticFile Is used to make data reusable and reduce I/O operations.
	StaticFile struct {
		Filename string
		Data     []byte
	}

	// Data structure used to create static home.json file.
	staticHome struct {
		NowPlayingPeriod scheduleutil.Period `json:"now_playing_week"`
		Movies           []StaticMovie       `json:"movies"`
	}

	// Data structure used to create static now_playing.json file.
	staticNowPlaying struct {
		Period scheduleutil.Period `json:"period"`
		Movies []StaticMovie       `json:"movies"`
	}
	weekPeriod struct {
		Start *time.Time `json:"start"`
		End   *time.Time `json:"end"`
	}

	// Data structure used to create static upcoming.json file.
	staticUpcoming []StaticMovie

	// Data structure used to retrieve now playing movie from database.
	nowPlayingMovie struct {
		// We need to add this `bson:",inline" to make fields from Movie be processed by mongo`
		models.Movie `bson:",inline"`
		Theaters     []models.Theater `bson:"cinemas"`
	}

	// StaticMovie represents a movie data structure, either now playing or upcoming,
	// used in creation of static files.
	StaticMovie struct {
		Title       string      `json:"title"`
		Poster      string      `json:"poster"`
		ReleaseDate *time.Time  `json:"release_date"`
		Theatres    *string     `json:"theatres"`
		MovieURL    string      `json:"movie_url"`
		SessionType SessionType `json:"session_type,omitempty"`
	}
)

// RunCreateStatic ...
func RunCreateStatic(t string) error {
	data, err := GetDataAccessLayer()
	if err != nil {
		return err
	}
	return CreateStatic(&Input{
		Name: "create_static",
		Args: []string{"-type", t},
	}, data)
}

// CreateStatic creates the static file correspoding to the given type for the given API.
func CreateStatic(in *Input, data persistence.DataAccessLayer) error {
	args := parseArgs(in)
	if args == nil {
		return errors.New("invalid arguments passed to create_static")
	}

	var err error

	st := ToStaticType(args["type"])
	switch st {
	default:
		// Do nothing.

	case StaticTypeHome:
		_, err = createStaticHome(data)
		break

	case StaticTypeNowPlaying:
		_, err = createStaticNowPlaying(data)
		break

	case StaticTypeUpcoming:
		_, err = createStaticUpcoming(data)
		break
	}

	return err
}

// ToStaticType converts a given string to its correspondent type
func ToStaticType(s string) StaticType {
	switch s {
	case "now_playing":
		return StaticTypeNowPlaying
	case "upcoming":
		return StaticTypeUpcoming
	}

	// Defaults to Home
	return StaticTypeHome
}

// Creates home.json static file.
// This requires an existent now_playing.json and upcoming.json files.
func createStaticHome(data persistence.DataAccessLayer) (*StaticFile, error) {
	// Make sure we have updated now_playing and upcoming JSON files.
	p1, err := createStaticNowPlaying(data)
	if err != nil {
		return nil, err
	}

	p2, err := createStaticUpcoming(data)
	if err != nil {
		return nil, err
	}

	var s1 staticNowPlaying
	var s2 staticUpcoming

	err = json.Unmarshal(p1.Data, &s1)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(p2.Data, &s2)
	if err != nil {
		return nil, err
	}

	h := staticHome{NowPlayingPeriod: s1.Period}
	h.Movies = append(s1.Movies, s2...)

	// TODO(diego):
	// Add to bucket with paths v1/home.json and home.json
	r, err := produceStaticFile("home.json", h)
	if err != nil {
		return nil, err
	}

	bucket := os.Getenv("STATIC_BUCKET_NAME")
	if bucket == "" {
		return nil, errors.New("missing bucket name")
	}

	err = UploadToBucket(bucket, "v1/"+r.Filename, r.Data)
	if err != nil {
		fmt.Println(err.Error())
	}
	err = UploadToBucket(bucket, r.Filename, r.Data)
	if err != nil {
		fmt.Println(err.Error())
	}

	return r, nil
}

// Creates now_playing.json static file for the v1 API.
func createStaticNowPlaying(data persistence.DataAccessLayer) (*StaticFile, error) {
	now := timeutil.Now()
	opts := &mongolayer.QueryOptions{
		Includes: []mongolayer.QueryInclude{
			mongolayer.QueryInclude{Field: "theaters"},
		},
	}
	movies, err := data.OldGetNowPlayingMovies(opts)
	if err != nil {
		return nil, err
	}

	period := scheduleutil.GetWeekPeriod(&now)
	result := staticNowPlaying{Period: *period}

	customSessionType := env.IsEnvVariableTrue("SESSION_TYPE_ENABLED")

	sm := make([]StaticMovie, 0)
	for _, movie := range movies {
		if movie.Hidden {
			continue
		}

		size := len(movie.Theaters)

		var theatres strings.Builder
		for i, cinema := range movie.Theaters {
			// NOTE(diego): Back-compat thing.
			// V1 API only supports IBICINEMAS and Cinemais Montes Claros.
			lower := strings.ToLower(cinema.Name)
			if strings.Contains(lower, "ibicinemas") || strings.Contains(lower, "cinemais montes claros") {
				theatres.WriteString(cinema.ShortName)
				if i < size-1 {
					theatres.WriteString(" - ")
				}
			}

		}
		tt := theatres.String()

		// NOTE(diego):
		// This code depends on movie release date to work correctly.
		// So... we need a crawler to keep these release dates updated for upcoming movies.
		//
		// 6 september 2018
		sessionType := SessionTypeNormal
		if customSessionType && movie.ReleaseDate != nil {
			maxDate := movie.ReleaseDate.AddDate(0, 0, 7)

			for _, session := range movie.Sessions {
				if session.StartTime.Before(*movie.ReleaseDate) && now.Before(*movie.ReleaseDate) {
					sessionType = SessionTypePreview
					break
				} else {
					// If session is exactly in the premiere week let's set this as premiere.
					if now.After(*movie.ReleaseDate) && now.Before(maxDate) {
						sessionType = SessionTypePremiere
						// NOTE(diego): Not breaking here because sessions may be in any order.
					}
				}
			}
		}

		sm = append(sm, StaticMovie{
			Title:       movie.Title,
			Poster:      applyAutoFormatForCloudinaryImage(movie.PosterURL),
			ReleaseDate: movie.ReleaseDate,
			SessionType: sessionType,
			Theatres:    &tt,
			MovieURL:    "/m/" + movie.ID.Hex(),
		})
	}

	result.Movies = sm

	r, err := produceStaticFile("movies/now_playing.json", result)
	if err != nil {
		return nil, err
	}

	bucket := os.Getenv("STATIC_BUCKET_NAME")
	if bucket == "" {
		return nil, errors.New("missing bucket name")
	}

	err = UploadToBucket(bucket, "v1/"+r.Filename, r.Data)
	if err != nil {
		fmt.Println(err.Error())
	}
	err = UploadToBucket(bucket, r.Filename, r.Data)
	if err != nil {
		fmt.Println(err.Error())
	}

	return r, nil
}

// Creates upcoming.json static file for the v1 API.
func createStaticUpcoming(data persistence.DataAccessLayer) (*StaticFile, error) {
	opts := &mongolayer.QueryOptions{Conditions: primitive.M{}}
	movies, err := data.GetUpcomingMovies(opts)
	if err != nil {
		return nil, err
	}

	s := make(staticUpcoming, len(movies))
	for index, movie := range movies {
		if movie.ID.Hex() == "" {
			continue
		}

		static := StaticMovie{
			Title:       movie.Title,
			Poster:      applyAutoFormatForCloudinaryImage(movie.PosterURL),
			ReleaseDate: movie.ReleaseDate,
			MovieURL:    "/m/" + movie.ID.Hex(),
		}

		s[index] = static
	}

	r, err := produceStaticFile("movies/upcoming.json", s)
	if err != nil {
		return nil, err
	}

	bucket := os.Getenv("STATIC_BUCKET_NAME")
	if bucket == "" {
		return nil, errors.New("missing bucket name")
	}

	err = UploadToBucket(bucket, "v1/"+r.Filename, r.Data)
	if err != nil {
		fmt.Println(err.Error())
	}
	err = UploadToBucket(bucket, r.Filename, r.Data)
	if err != nil {
		fmt.Println(err.Error())
	}

	return r, nil
}

func produceStaticFile(filename string, data interface{}) (*StaticFile, error) {
	if filename == "" {
		return nil, errors.New("filename is missing")
	}

	if data == nil {
		return nil, errors.New("data is missing")
	}

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &StaticFile{filename, b}, nil
}

// NOTE: isolate to another file
func applyAutoFormatForCloudinaryImage(url string) string {
	base := "res.cloudinary.com/dyrib46is/image/upload"
	index := strings.Index(url, base)
	if index > -1 {
		baseLen := len(base)
		start := url[0 : index+baseLen]
		end := url[index+baseLen:]
		return start + "/f_auto" + end
	}

	return url
}

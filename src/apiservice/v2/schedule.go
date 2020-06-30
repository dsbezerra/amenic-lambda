package v2

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dsbezerra/amenic-lambda/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/models"
	"github.com/dsbezerra/amenic-lambda/src/lib/util/apiutil"
	"github.com/dsbezerra/amenic-lambda/src/lib/util/timeutil"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ScheduleService ...
type ScheduleService struct {
	data persistence.DataAccessLayer
}

// Room ...
type Room struct {
	ID         string          `json:"-"`
	Number     uint            `json:"number" bson:"number"`
	Name       string          `json:"name" bson:"name"`
	Attributes []RoomAttribute `json:"attributes"`
	Sessions   []Session       `json:"sessions" bson:"sessions"`
}

// RoomAttribute ...
type RoomAttribute struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
}

// Schedule ...
type Schedule struct {
	Rooms   []Room          `json:"rooms" bson:"rooms"`
	Movie   *models.Movie   `json:"movie,omitempty" bson:"movie,omitempty"`
	Theater *models.Theater `json:"theater,omitempty" bson:"theater,omitempty"`
}

// Session ...
type Session struct {
	ID        string `json:"_id"`
	Active    bool   `json:"active"`
	StartTime string `json:"startTime"`
}

// ServeSchedules ...
func (r *RESTService) ServeSchedules(rg *gin.RouterGroup) {
	s := &ScheduleService{r.data}

	schedules := rg.Group("/schedules", rest.JWTAuth(nil))
	schedules.GET("", s.GetAll)
}

// GetAll ...
func (s *ScheduleService) GetAll(c *gin.Context) {
	query := BuildScheduleQuery(s.data, c)
	if query == nil {
		apiutil.SendBadRequest(c)
		return
	}

	sessions, err := s.data.GetSessions(query)
	if err != nil {
		apiutil.SendSuccessOrError(c, nil, err)
		return
	}
	schedules := mapSessionsToSchedules(sessions)
	apiutil.SendSuccessOrError(c, schedules, err)
}

// BuildScheduleQuery ...
func BuildScheduleQuery(data persistence.DataAccessLayer, c *gin.Context) persistence.Query {
	// @Refactor to support other databases?
	qopts := c.MustGet("query_options").(map[string]string)
	theaterID, err := primitive.ObjectIDFromHex(qopts["theaterId"])
	if err != nil {
		return nil
	}
	query := data.DefaultQuery().
		AddCondition("hidden", false).
		AddCondition("theaterId", theaterID)

	movieID, err := primitive.ObjectIDFromHex(qopts["movieId"])
	if err == nil {
		query.AddCondition("movieId", movieID)
	}

	loc, _ := time.LoadLocation("America/Sao_Paulo")
	t, err := time.ParseInLocation("2006-01-02", qopts["date"], loc)
	if err != nil {
		t = timeutil.StartOfDay()
	}
	d := fmt.Sprintf("%d%02d%02d", t.Year(), int(t.Month()), t.Day())
	v, _ := strconv.Atoi(d)
	query.AddCondition("date", v)
	query.AddInclude("movie")
	query.SetSort("movie.title", "room", "version", "format", "startTime")
	query.SetLimit(-1)
	return query
}

func makeRoom(session *models.Session) Room {
	var attrs []RoomAttribute

	attr := RoomAttribute{}
	if session.Format == models.Format2D {
		attr.ID = models.Format2D
		attr.Name = "2D"
	} else if session.Format == models.Format3D {
		attr.ID = models.Format3D
		attr.Name = "3D"
	}

	if attr.ID != "" {
		attrs = append(attrs, attr)
	}

	attr = RoomAttribute{}
	if session.Version == models.VersionDubbed {
		attr.ID = models.VersionDubbed
		attr.Name = "Dublado"
	} else if session.Version == models.VersionSubbed || session.Version == models.VersionSubtitled {
		attr.ID = models.VersionSubtitled
		attr.Name = "Legendado"
	} else if session.Version == models.VersionNational {
		attr.ID = models.VersionNational
		attr.Name = "Nacional"
	}

	if attr.ID != "" {
		attrs = append(attrs, attr)
	}

	ID := strings.Builder{}
	ID.WriteString(fmt.Sprintf("%d-", session.Room))
	for _, a := range attrs {
		ID.WriteString(a.ID)
	}

	active := time.Now().In(time.UTC).Before(*session.StartTime)
	return Room{
		ID:         ID.String(),
		Name:       fmt.Sprintf("Sala %d", session.Room),
		Number:     session.Room,
		Attributes: attrs,
		Sessions: []Session{
			Session{
				ID:        session.ID.Hex(),
				Active:    active,
				StartTime: session.OpeningTime,
			},
		},
	}
}

func mapSessionsToSchedules(sessions []models.Session) []Schedule {

	var schedules []Schedule
	if sessions == nil || len(sessions) == 0 {
		return schedules
	}

	now := time.Now().In(time.UTC)
	schedmap := map[string]int{}

	for _, session := range sessions {
		if session.Hidden || session.Room == 0 {
			continue
		}

		index, ok := schedmap[session.MovieID.Hex()]
		if !ok {
			index = len(schedules)
			schedules = append(schedules, Schedule{
				Movie: session.Movie,
			})
			schedmap[session.MovieID.Hex()] = index
		}

		room := makeRoom(&session)

		found := -1
		for i, r := range schedules[index].Rooms {
			if r.ID == room.ID {
				room = r
				found = i
				break
			}
		}

		if found == -1 {
			schedules[index].Rooms = append(schedules[index].Rooms, room)
		} else {
			room.Sessions = append(room.Sessions, Session{
				ID:        session.ID.Hex(),
				StartTime: session.OpeningTime,
				Active:    now.Before(*session.StartTime),
			})
			schedules[index].Rooms[found] = room
		}
	}

	return schedules
}

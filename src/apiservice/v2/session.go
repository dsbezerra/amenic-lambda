package v2

import (
	"time"

	"github.com/dsbezerra/amenic-lambda/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/dsbezerra/amenic-lambda/src/lib/util/apiutil"
	"github.com/dsbezerra/amenic-lambda/src/lib/util/scheduleutil"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// SessionService ...
type SessionService struct {
	data persistence.DataAccessLayer
}

// ServeSessions ...
func (r *RESTService) ServeSessions(rg *gin.RouterGroup) {
	s := &SessionService{r.data}

	client := rg.Group("/sessions", rest.JWTAuth(nil))
	client.GET("/session/:id", s.Get)

	admin := rg.Group("/sessions", rest.JWTAuth(&rest.Endpoint{AdminOnly: true}))
	admin.GET("", s.GetAll)
}

// Get gets the session corresponding the requested ID.
func (s *SessionService) Get(c *gin.Context) {
	session, err := s.data.GetSession(c.Param("id"), BuildSessionQuery(s.data, c))
	apiutil.SendSuccessOrError(c, session, err)
}

// GetAll gets all sessions.
func (s *SessionService) GetAll(c *gin.Context) {
	q := c.MustGet("query_options").(map[string]string)

	query := s.data.BuildSessionQuery(q)

	var hasStart, hasEnd bool
	start, ok := q["start"]
	if ok {
		t, err := time.ParseInLocation("2006-01-02", start, time.UTC)
		if err == nil {
			query.AddCondition("startTime", bson.M{"$gte": t})
			hasStart = true
		}
	}

	end, ok := q["end"]
	if ok {
		t, err := time.ParseInLocation("2006-01-02", end, time.UTC)
		if err == nil {
			t.Add((time.Hour * 24) - time.Nanosecond)
			query.AddCondition("startTime", bson.M{"$lte": t})
			hasEnd = true
		}
	}

	if hasStart && hasEnd {
		query.SetLimit(-1)
	} else if !hasStart {
		period := scheduleutil.GetWeekPeriod(nil)
		query.AddCondition("startTime", bson.M{"$gte": period.Start})
		query.SetLimit(-1)
	}

	// If we are not sorting let's set the default sort
	if !query.Sorting() {
		query.SetSort("-movieId", "+version", "+format", "+startTime")
	}

	sessions, err := s.data.GetSessions(query)
	apiutil.SendSuccessOrError(c, sessions, err)
}

// BuildSessionQuery builds session query from request query string
func BuildSessionQuery(data persistence.DataAccessLayer, c *gin.Context) persistence.Query {
	query := c.MustGet("query_options").(map[string]string)
	return data.BuildSessionQuery(query)
}

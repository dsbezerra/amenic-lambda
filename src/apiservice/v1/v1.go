/** DEPRECATED */

package v1

import (
	"github.com/dsbezerra/amenic-lambda/src/lib/middlewares"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/gin-gonic/gin"
)

type (
	// RESTService TODO
	RESTService struct {
		data persistence.DataAccessLayer
	}
)

// AddRoutes add V1 routes to main router in group v1
func AddRoutes(r *gin.Engine, data persistence.DataAccessLayer) {
	r.Use(middlewares.BaseParseQuery())
	v1 := r.Group("v1")
	s := RESTService{data}
	s.ServeAPIKeys(v1)
	s.ServeMovies(v1)
	s.ServeNotifications(v1)
	s.ServePrices(v1)
	s.ServeScores(v1)
	s.ServeShowtimes(v1)
	s.ServeTheaters(v1)
}

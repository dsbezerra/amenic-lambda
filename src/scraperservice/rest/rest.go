package rest

import (
	"github.com/dsbezerra/amenic-lambda/src/lib/messagequeue"
	"github.com/dsbezerra/amenic-lambda/src/lib/middlewares"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/gin-gonic/gin"
)

// Service TODO
type Service struct {
	data    persistence.DataAccessLayer
	emitter messagequeue.EventEmitter
}

// ServeAPI ...
func ServeAPI(r *gin.Engine, data persistence.DataAccessLayer, emitter messagequeue.EventEmitter) {
	s := &Service{data, emitter}

	// Apply default middlewares
	r.Use(middlewares.BaseParseQuery())

	// ScraperService routes.
	s.ServeScrapers(r)
}

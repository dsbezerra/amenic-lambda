package v2

import (
	"github.com/dsbezerra/amenic-lambda/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/models"
	"github.com/dsbezerra/amenic-lambda/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
)

// StateResponse ...
type StateResponse struct {
	ID   models.State `json:"_id"`
	Name string       `json:"name"`
}

// StateService ...
type StateService struct {
	data persistence.DataAccessLayer
}

// ServeStates ...
func (r *RESTService) ServeStates(rg *gin.RouterGroup) {
	s := &StateService{r.data}

	states := rg.Group("/states", rest.JWTAuth(&rest.Endpoint{AdminOnly: true}))
	states.GET("", s.GetAll)
	states.GET("/state/:id", s.Get)
	states.GET("/state/:id/cities", s.GetCities)
	// TODO: Get theaters
}

// Get gets the State corresponding the requested ID.
func (s *StateService) Get(c *gin.Context) {
	state, ok := models.GetState(c.Param("id"))
	if ok {
		apiutil.SendSuccessOrError(c, StateResponse{ID: state, Name: state.Name()}, nil)
		return
	}
	apiutil.SendBadRequest(c)
}

// GetAll gets all States.
func (s *StateService) GetAll(c *gin.Context) {
	states := []StateResponse{}
	for _, state := range models.GetStateList() {
		states = append(states, StateResponse{ID: state, Name: state.Name()})
	}
	apiutil.SendSuccessOrError(c, states, nil)
}

// GetCities gets all cities from the given State.
func (s *StateService) GetCities(c *gin.Context) {
	state, ok := models.GetState(c.Param("id"))
	if !ok {
		apiutil.SendBadRequest(c)
		return
	}
	query := c.MustGet("query_options").(map[string]string)
	query["state"] = string(state)
	cities, err := s.data.GetCities(s.data.BuildCityQuery(query))
	apiutil.SendSuccessOrError(c, cities, err)
}

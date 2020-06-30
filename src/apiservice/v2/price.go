package v2

import (
	"github.com/dsbezerra/amenic-lambda/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/dsbezerra/amenic-lambda/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
)

// PriceService ...
type PriceService struct {
	data persistence.DataAccessLayer
}

// ServePrices ...
func (r *RESTService) ServePrices(rg *gin.RouterGroup) {
	s := &PriceService{r.data}

	client := rg.Group("/prices", rest.JWTAuth(nil))
	client.GET("/price/:id", s.Get)

	admin := rg.Group("/prices", rest.JWTAuth(&rest.Endpoint{AdminOnly: true}))
	admin.GET("", s.GetAll)
}

// Get gets the price corresponding the requested ID.
func (s *PriceService) Get(c *gin.Context) {
	price, err := s.data.GetPrice(c.Param("id"), BuildPriceQuery(s.data, c))
	apiutil.SendSuccessOrError(c, price, err)
}

// GetAll gets all prices.
func (s *PriceService) GetAll(c *gin.Context) {
	prices, err := s.data.GetPrices(BuildPriceQuery(s.data, c))
	apiutil.SendSuccessOrError(c, prices, err)
}

// BuildPriceQuery builds price query from request query string
func BuildPriceQuery(data persistence.DataAccessLayer, c *gin.Context) persistence.Query {
	query := c.MustGet("query_options").(map[string]string)
	return data.BuildPriceQuery(query)
}

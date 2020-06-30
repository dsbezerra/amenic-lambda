/** DEPRECATED */

package v1

import (
	"time"

	"github.com/dsbezerra/amenic-lambda/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/models"
	"github.com/dsbezerra/amenic-lambda/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	Price struct {
		ID         primitive.ObjectID `json:"_id" bson:"_id"`
		CinemaID   string             `json:"cinema_id" bson:"cinema_id"`
		Label      string             `json:"label" bson:"label"`
		Amount     uint               `json:"amount" bson:"amount"`
		Is2D       bool               `json:"is_2D" bson:"is_2D"` // TODO(diego): This should be removed once all apps are over v1.1.1
		Weekdays   []models.Weekday   `json:"weekdays" bson:"weekdays"`
		Exceptions []models.Weekday   `json:"exceptions" bson:"exceptions"`
		Attributes []string           `json:"attributes" bson:"attributes"`
		Weight     uint               `json:"-" bson:"weight"` // NOTE(diego): Used only to sort
	}
)

// PriceService ...
type PriceService struct {
	data persistence.DataAccessLayer
}

func (r *RESTService) ServePrices(rg *gin.RouterGroup) {
	s := &PriceService{r.data}
	prices := rg.Group("/prices", rest.ClientAuth(r.data))
	prices.GET("", s.GetAll)
}

func (s *PriceService) GetAll(c *gin.Context) {
	cinemaQuery := s.data.DefaultQuery()
	if cinemaID := c.Query("cinema"); cinemaID != "" {
		if cinemaID == "ibicinemas" {
			cinemaQuery.AddCondition("internalId", cinemaID)
		} else if cinemaID == "cinemais-34" {
			cinemaQuery.AddCondition("internalId", "34").AddCondition("shortName", "Cinemais")
		} else {
			apiutil.SendBadRequest(c)
			return
		}
	}

	cinema, err := s.data.FindTheater(cinemaQuery)
	if err != nil {
		apiutil.SendSuccessOrError(c, nil, err)
		return
	}

	prices, err := s.data.GetPrices(s.data.DefaultQuery().AddCondition("theaterId", cinema.ID))
	apiutil.SendSuccessOrError(c, s.mapTo(prices, cinema.InternalID), err)
}

func (s *PriceService) mapTo(prices []models.Price, cinema string) []Price {
	if prices == nil || (cinema != "ibicinemas" && cinema != "34") {
		return nil
	}

	contains := func(arr []string, str string) bool {
		for _, s := range arr {
			if s == str {
				return true
			}
		}
		return false
	}

	weekdays := func(arr []time.Weekday) []models.Weekday {
		var result []models.Weekday
		for _, w := range arr {
			r := models.TimeWeekdayToWeekday(w)
			if r != models.INVALID {
				result = append(result, r)
			}
		}
		return result
	}

	var result []Price
	for _, p := range prices {
		price := Price{
			ID:         p.ID,
			CinemaID:   cinema,
			Label:      p.Label,
			Amount:     uint(p.Full),
			Is2D:       contains(p.Attributes, "2D"),
			Weekdays:   weekdays(p.Weekdays),
			Attributes: p.Attributes,
			Weight:     p.Weight,
		}
		if p.IncludingHolidays {
			price.Weekdays = append(price.Weekdays, models.HOLIDAY)
		}
		if p.ExceptHolidays {
			price.Exceptions = append(price.Exceptions, models.HOLIDAY)
		}
		result = append(result, price)
	}

	return result
}

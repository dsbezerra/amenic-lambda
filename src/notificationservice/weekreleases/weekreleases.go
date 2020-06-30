package weekreleases

import (
	"fmt"
	"strings"
	"time"

	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// Type ...
	Type = "week_releases"

	// Title ...
	Title = "Estreias da semana"
)

// PrepareNotification ...
func PrepareNotification(releases []models.Movie) *models.Notification {
	if len(releases) == 0 {
		return nil
	}

	notification := &models.Notification{
		ID:        primitive.NewObjectID(),
		Type:      Type,
		Title:     Title,
		CreatedAt: time.Now().UTC(),
	}

	text := strings.Builder{}
	htmlText := strings.Builder{}

	size := len(releases)

	for mIndex, m := range releases {
		theaters := strings.Builder{}

		cSize := len(m.Theaters)
		if cSize == 1 {
			theaters.WriteString(m.Theaters[0].Name)
		} else if cSize > 1 {
			// NOTE(diego): Initially we add all theaters here, but *WHEN AND IF* we support more theaters
			// we need to refactor this.
			// In that case it would be better to just create FCM topics for each theater and run some sort of routine that
			// checks the week releases for each theater and if we find a movie notify user.
			for cIndex, c := range m.Theaters {
				theaters.WriteString(c.Name)
				if cIndex < cSize-1 {
					theaters.WriteString(", ")
				}
			}
		}

		var t string
		var h string
		if cSize == 0 {
			t = fmt.Sprintf("%s", m.Title)
			h = fmt.Sprintf("<b>%s</b>", m.Title)
		} else {
			t = fmt.Sprintf("%s (%s)", m.Title, theaters.String())
			h = fmt.Sprintf("<b>%s</b> (<i>%s</i>)", m.Title, theaters.String())
		}

		text.WriteString(t)
		htmlText.WriteString(h)

		if mIndex < size-1 {
			text.WriteString("\n")
			htmlText.WriteString("<br/>")
		}

		if notification.Data == nil {
			notification.Data = &models.NotificationData{}
		}

		notification.Data.Movies = append(notification.Data.Movies, models.Movie{
			ID:            m.ID,
			Title:         m.Title,
			OriginalTitle: m.OriginalTitle,
			PosterURL:     m.PosterURL,
			Genres:        m.Genres,
			Runtime:       m.Runtime,
			Trailer:       m.Trailer,
		})
	}

	notification.Single = len(notification.Data.Movies) == 1
	if notification.Single {
		notification.ItemID = notification.Data.Movies[0].ID.Hex()
	}

	notification.Text = text.String()
	notification.HTMLText = htmlText.String()

	return notification
}

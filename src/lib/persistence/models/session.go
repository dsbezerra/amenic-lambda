package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	Format2D = "2D"
	Format3D = "3D"

	VersionSubbed    = "subbed"
	VersionSubtitled = "subtitled"
	VersionDubbed    = "dubbed"
	VersionNational  = "national"
)

// Session ...
type Session struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	MovieID     primitive.ObjectID `json:"movieId,omitempty" bson:"movieId,omitempty"`
	TheaterID   primitive.ObjectID `json:"theaterId,omitempty" bson:"theaterId,omitempty"`
	MovieSlugs  Slugs              `json:"movieSlugs,omitempty" bson:"movieSlugs,omitempty"`
	Hidden      bool               `json:"hidden" bson:"hidden"`
	Format      string             `json:"format" bson:"format"`
	Version     string             `json:"version" bson:"version"`
	Room        uint               `json:"room" bson:"room"`
	TimeZone    string             `json:"timeZone,omitempty" bson:"timeZone,omitempty"`
	OpeningTime string             `json:"openingTime,omitempty" bson:"openingTime,omitempty"`
	Date        int                `json:"date,omitempty" bson:"date,omitempty"`
	StartTime   *time.Time         `json:"startTime" bson:"startTime"`
	CreatedAt   *time.Time         `json:"createdAt,omitempty" bson:"createdAt"`
	Theater     *Theater           `json:"theater,omitempty" bson:"theater,omitempty"`
	Movie       *Movie             `json:"movie,omitempty" bson:"movie,omitempty"`
}

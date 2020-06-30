package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// These flags are used to lock members of Movie struct and prevent its modification
// by any scraper
const (
	// MovieLockOriginalTitle locks OriginalTitle
	MovieLockOriginalTitle uint64 = 1 << iota
	// MovieLockTitle locks Title
	MovieLockTitle
	// MovieLockSynopsis locks Synopsis
	MovieLockSynopsis
	// TODO: Add more as necessary
)

// Movie represents a movie
type Movie struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Hidden        bool               `json:"hidden,omitempty" bson:"hidden"` // Whether this movie should be retrieved or not.
	ClaqueteID    int                `json:"claquete_id,omitempty" bson:"claqueteId,omitempty"`
	TmdbID        int                `json:"tmdb_id,omitempty" bson:"tmdbId,omitempty"`
	ImdbID        string             `json:"imdb_id,omitempty" bson:"imdbId,omitempty"`
	Slugs         Slugs              `json:"-" bson:"slugs,omitempty"`
	Title         string             `json:"title,omitempty" bson:"title,omitempty"`
	OriginalTitle string             `json:"original_title,omitempty" bson:"originalTitle,omitempty"`
	Cast          []string           `json:"cast,omitempty" bson:"cast,omitempty"`
	PosterURL     string             `json:"poster_url,omitempty" bson:"poster,omitempty"`
	BackdropURL   string             `json:"backdrop_url,omitempty" bson:"backdrop,omitempty"`
	Synopsis      string             `json:"synopsis,omitempty" bson:"synopsis,omitempty"`
	Trailer       string             `json:"trailer_id,omitempty" bson:"trailer,omitempty"`
	Genres        []string           `json:"genres,omitempty" bson:"genres,omitempty"`
	Rating        int                `json:"rating,omitempty" bson:"rating,omitempty"`
	Runtime       int                `json:"runtime,omitempty" bson:"runtime,omitempty"`
	Distributor   string             `json:"distributor,omitempty" bson:"studio,omitempty"`
	ReleaseDate   *time.Time         `json:"release_date,omitempty" bson:"releaseDate,omitempty"`
	CreatedAt     *time.Time         `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt     *time.Time         `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	Theaters      []Theater          `json:"cinemas,omitempty" bson:"theaters,omitempty"`
	Scores        []Score            `json:"scores,omitempty" bson:"scores,omitempty"`
	Sessions      []Session          `json:"sessions,omitempty" bson:"sessions,omitempty"`
	LockFlags     uint64             `json:"-" bson:"lockFlags,omitempty"`
}

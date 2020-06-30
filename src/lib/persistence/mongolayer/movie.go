package mongolayer

import (
	"context"
	"fmt"
	"strconv"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/models"
	"github.com/dsbezerra/amenic-lambda/src/lib/util/scheduleutil"
	"github.com/dsbezerra/amenic-lambda/src/lib/util/timeutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CountMovies ...
func (m *MongoDAL) CountMovies(query persistence.Query) (int64, error) {
	return m.C(CollectionMovies).CountDocuments(context.Background(), query.GetConditions())
}

// InsertMovie ...
func (m *MongoDAL) InsertMovie(movie models.Movie) error {
	_, err := m.C(CollectionMovies).InsertOne(context.Background(), movie)
	return err
}

// FindMovie ...
func (m *MongoDAL) FindMovie(query persistence.Query) (*models.Movie, error) {
	var result models.Movie

	var ctx = context.Background()
	var cursor *mongo.Cursor
	var err error

	var C = m.C(CollectionMovies)
	if query.HasInclude() {
		cursor, err = C.Aggregate(ctx, buildPipeline(CollectionMovies, query.(*QueryOptions)))
		if cursor != nil {
			defer cursor.Close(ctx)
			if cursor.Next(ctx) {
				err = cursor.Decode(&result)
			}
		}
	} else {
		err = C.FindOne(ctx, query.GetConditions(), getFindOneOptions(query)).Decode(&result)
	}

	if err != nil {
		return nil, err
	}

	return &result, err
}

// FindMovieAndUpdate finds a Movie matching the query and updates it, returning either the original or
// updated.
func (m *MongoDAL) FindMovieAndUpdate(query persistence.Query, update interface{}) (*models.Movie, error) {
	var result models.Movie
	err := m.C(CollectionMovies).FindOneAndUpdate(context.Background(), query.GetConditions(), update, getFindOneAndUpdateOptions(query)).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// GetMovie ...
func (m *MongoDAL) GetMovie(id string, query persistence.Query) (*models.Movie, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return m.FindMovie(query.AddCondition("_id", ID))
}

// GetMovies ...
func (m *MongoDAL) GetMovies(query persistence.Query) ([]models.Movie, error) {
	// TODO: Implement aggregate for include queries
	var result = []models.Movie{}
	var ctx = context.Background()
	cursor, err := m.C(CollectionMovies).Find(ctx, query.GetConditions(), getFindOptions(query))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &result)
	return result, err
}

// GetMoviesByTitle ...
func (m *MongoDAL) GetMoviesByTitle(title string) ([]models.Movie, error) {
	opts := DefaultOptions("").
		AddCondition("$text", bson.M{"$search": title}).
		SetFields(bson.M{
			"score": bson.M{"$meta": "textScore"},
		}).
		SetSort("$testScore:score")
	return m.GetMovies(opts)
}

// GetNowPlayingMovies returns now playing movies for the given query condition
func (m *MongoDAL) GetNowPlayingMovies(query persistence.Query) ([]models.Movie, error) {

	var result = []models.Movie{}

	// NOTE(diego):
	// The query argument is mostly expected to be filled by BuildSessionQuery
	// function called in /movies/now_playing route.
	//
	// However, we still have some checks to make sure it return the current
	// now playing movies in case the query instance is empty.
	var opts *QueryOptions
	if query == nil {
		opts = &QueryOptions{}
	} else {
		opts = query.(*QueryOptions)
	}

	// Builds $and match property
	and := []bson.M{}
	for name, value := range opts.Conditions {
		and = append(and, bson.M{name: value})
	}

	// If we don't have conditions let's return only the current
	// now playing movies
	if len(and) == 0 {
		period := scheduleutil.GetWeekPeriod(nil)
		and = append(and, bson.M{"startTime": bson.M{"$gte": period.Start}})
	}

	var includeTheaters bool
	if query.HasInclude() {
		includes := query.GetIncludes().([]QueryInclude)
		for _, inc := range includes {
			if inc.Field == "theaters" {
				includeTheaters = true
				break
			}
		}
	}

	// Pipeline begin by finding all sessions, matching the given conditions, and grouping them by
	// movie. Therefore we have a lookup stage to complete movie data.
	p := []bson.M{
		{
			"$match": bson.M{
				"$and": and,
			},
		},
	}

	if includeTheaters {
		p = append(p,
			bson.M{
				"$group": bson.M{
					"_id": "$movieId",
					"theaters": bson.M{
						"$addToSet": "$theaterId",
					},
				},
			})
	} else {
		p = append(p, bson.M{
			"$group": bson.M{
				"_id": "$movieId",
			},
		})
	}

	p = append(p, bson.M{
		"$lookup": bson.M{
			"from":         CollectionMovies,
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "movie",
		},
	})

	if includeTheaters {
		p = append(p, bson.M{
			"$lookup": bson.M{
				"from":         CollectionTheaters,
				"localField":   "theaters",
				"foreignField": "_id",
				"as":           "theaters",
			},
		})
	}

	p = append(p, bson.M{"$unwind": "$movie"})

	if includeTheaters {
		p = append(p, bson.M{
			"$addFields": bson.M{
				"movie.theaters": "$theaters",
			},
		})
	}

	// Builds the $sort stage which sorts everything by release date in
	// descending order unless we define a custom sort.
	sort := bson.D{}
	if opts.sorting {
		rest := opts.Sort[0:]
		opts.Sort = []string{"-movie.releaseDate"}
		opts.Sort = append(opts.Sort, rest...)
		sort = SortToBSON("movie", opts.Sort...)
		if len(sort) > 0 {
			p = append(p, bson.M{"$sort": sort})
		}
	}

	// Now we add the final stages, which are:
	project := bson.M{}

	if len(opts.Fields) == 0 {
		opts.Fields = primitive.M{
			"_id":         1,
			"poster":      1,
			"title":       1,
			"trailer":     1,
			"rating":      1,
			"releaseDate": 1,
		}
	}

	for f := range opts.Fields {
		field := fmt.Sprintf("movie.%s", f)
		project[field] = 1
	}

	if includeTheaters {
		project["movie.theaters"] = 1
	}

	// project our desired fields to the final result
	p = append(p, bson.M{"$project": project})

	// will make sure ou movie documents are in the root
	p = append(p, bson.M{"$replaceRoot": bson.M{"newRoot": "$movie"}})

	var ctx = context.Background()
	cursor, err := m.C(CollectionSessions).Aggregate(ctx, p)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &result)
	return result, err
}

// OldGetNowPlayingMovies retrieves all now playing movies for the given conditions.
// **************   FOR STATIC home.json AND now_playing.json FILES   ****************
func (m *MongoDAL) OldGetNowPlayingMovies(query persistence.Query) ([]models.Movie, error) {
	var result = []models.Movie{}

	period := scheduleutil.GetWeekPeriod(nil)

	pipe := []bson.M{
		{
			"$match": bson.M{"startTime": bson.M{"$gte": period.Start}},
		},
		{
			"$group": bson.M{
				"_id": "$movieId",
				"theaters": bson.M{
					"$addToSet": "$theaterId",
				},
			},
		},
		{
			"$lookup": bson.M{
				"from":         CollectionMovies,
				"localField":   "_id",
				"foreignField": "_id",
				"as":           "movie",
			},
		},
		{
			"$lookup": bson.M{
				"from":         CollectionTheaters,
				"localField":   "theaters",
				"foreignField": "_id",
				"as":           "theaters",
			},
		},
		{
			"$unwind": "$movie", //  is used to remove the movie from the returned array
		},
		{
			"$addFields": bson.M{
				"movie.theaters": "$theaters",
			},
		},
		{
			"$project": bson.M{
				"movie._id":         1,
				"movie.title":       1,
				"movie.poster":      1,
				"movie.releaseDate": 1,
				"movie.theaters":    1,
			},
		},
		{
			"$replaceRoot": bson.M{
				"newRoot": "$movie",
			},
		},
	}

	var ctx = context.Background()
	cursor, err := m.C(CollectionSessions).Aggregate(ctx, pipe)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &result)
	return result, err
}

// GetUpcomingMovies ...
func (m *MongoDAL) GetUpcomingMovies(query persistence.Query) ([]models.Movie, error) {
	// Retrieve start of the current day timestamp
	startOfDay := timeutil.StartOfDay()
	// Builds default query for this operation
	opts := query.(*QueryOptions)
	opts.
		AddCondition("hidden", false).
		AddCondition("releaseDate", bson.M{"$gt": startOfDay})
	// Set default sort if we don't specify one
	if !opts.sorting {
		opts.Sort = []string{"+releaseDate"}
	}
	if len(opts.Fields) == 0 {
		opts.Fields = primitive.M{
			"_id":         1,
			"poster":      1,
			"backdrop":    1,
			"title":       1,
			"trailer":     1,
			"rating":      1,
			"releaseDate": 1,
		}
	}
	return m.GetMovies(opts)
}

// UpdateMovie ...
func (m *MongoDAL) UpdateMovie(id string, mm models.Movie) (int64, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, err
	}
	mm.UpdatedAt = getCurrentTime()
	result, err := m.C(CollectionMovies).UpdateOne(context.Background(), bson.M{"_id": ID}, bson.M{"$set": mm})
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, err
}

// DeleteMovie ...
func (m *MongoDAL) DeleteMovie(id string) error {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.C(CollectionMovies).DeleteOne(context.Background(), bson.M{"_id": ID})
	return err
}

// DeleteMovies ...
func (m *MongoDAL) DeleteMovies(query persistence.Query) (int64, error) {
	result, err := m.C(CollectionMovies).DeleteMany(context.Background(), query.GetConditions())
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, err
}

// BuildMovieQuery converts a map of query string to mongolayer syntax for Movie model
func (m *MongoDAL) BuildMovieQuery(q map[string]string) persistence.Query {
	query := BuildQuery("", q)
	if len(q) > 0 {
		// IDs
		ID, ok := q["id"]
		if ok {
			value, err := primitive.ObjectIDFromHex(ID)
			if err == nil {
				query.AddCondition("_id", value)
			}
		}
		ID, ok = q["claqueteId"]
		if ok {
			value, err := strconv.Atoi(ID)
			if err == nil {
				query.AddCondition("claqueteId", value)
			}
		}
		ID, ok = q["tmdbId"]
		if ok {
			value, err := strconv.Atoi(ID)
			if err == nil {
				query.AddCondition("tmdbId", value)
			}
		}
		ID, ok = q["imdbId"]
		if ok {
			query.AddCondition("imdbId", ID)
		}

		// Hidden
		hidden, ok := q["hidden"]
		if ok {
			value, err := strconv.ParseBool(hidden)
			if err == nil {
				query.AddCondition("hidden", value)
			}
		}

		backdrop, ok := q["backdrop"]
		if ok {
			value, err := strconv.ParseBool(backdrop)
			if err == nil {
				prop := "$eq"
				if value {
					prop = "$ne"
				}
				query.AddCondition("backdrop", bson.M{prop: ""})
			}
		}

		poster, ok := q["poster"]
		if ok {
			value, err := strconv.ParseBool(poster)
			if err == nil {
				prop := "$eq"
				if value {
					prop = "$ne"
				}
				query.AddCondition("poster", bson.M{prop: ""})
			}
		}

		trailer, ok := q["trailer"]
		if ok {
			value, err := strconv.ParseBool(trailer)
			if err == nil {
				prop := "$eq"
				if value {
					prop = "$ne"
				}
				query.AddCondition("trailer", bson.M{prop: ""})
			}
		}

		rating, ok := q["rating"]
		if ok {
			value, err := strconv.Atoi(rating)
			if err == nil && value == -1 || value >= 10 && value <= 18 {
				query.AddCondition("rating", value)
			}
		}

		search, ok := q["search"]
		if ok {
			query.AddCondition("$or", []bson.M{
				bson.M{"title": bson.M{"$regex": search, "$options": "ig"}},
				bson.M{"originalTitle": bson.M{"$regex": search, "$options": "ig"}},
			})
		}
	}
	return query
}

package repositories

import (
	"context"
	"errors"

	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrMovieNotFound      = errors.New("movie not found")
	ErrMovieAlreadyExists = errors.New("movie already exists")
)

// MovieRepository defines the interface for movie data operations
type MovieRepository interface {
	Create(ctx context.Context, movie *models.Movie) error
	FindAll(ctx context.Context, filter bson.M, opts []*options.FindOptions) ([]models.Movie, error)
	FindByID(ctx context.Context, id string) (*models.Movie, error)
	FindByImdbID(ctx context.Context, imdbID string) (*models.Movie, error)
	FindByGenre(ctx context.Context, genreID int, limit, skip int) ([]models.Movie, error)
	FindByGenres(ctx context.Context, genreIDs []int, limit int) ([]models.Movie, error)
	Update(ctx context.Context, id string, update bson.M) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context, filter bson.M) (int64, error)
	MovieExists(ctx context.Context, imdbID string) (bool, error)
}

// movieRepositoryImpl implements MovieRepository
type movieRepositoryImpl struct {
	collection *mongo.Collection
}

// NewMovieRepository creates a new movie repository
func NewMovieRepository(collection *mongo.Collection) MovieRepository {
	return &movieRepositoryImpl{
		collection: collection,
	}
}

func (r *movieRepositoryImpl) Create(ctx context.Context, movie *models.Movie) error {
	// Check if movie already exists
	exists, err := r.MovieExists(ctx, movie.ImdbID)
	if err != nil {
		return err
	}
	if exists {
		return ErrMovieAlreadyExists
	}

	_, err = r.collection.InsertOne(ctx, movie)
	return err
}

func (r *movieRepositoryImpl) FindAll(ctx context.Context, filter bson.M, opts []*options.FindOptions) ([]models.Movie, error) {
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var movies []models.Movie
	if err := cursor.All(ctx, &movies); err != nil {
		return nil, err
	}

	return movies, nil
}

func (r *movieRepositoryImpl) FindByID(ctx context.Context, id string) (*models.Movie, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var movie models.Movie
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&movie)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrMovieNotFound
		}
		return nil, err
	}

	return &movie, nil
}

func (r *movieRepositoryImpl) FindByImdbID(ctx context.Context, imdbID string) (*models.Movie, error) {
	var movie models.Movie
	err := r.collection.FindOne(ctx, bson.M{"imdb_id": imdbID}).Decode(&movie)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrMovieNotFound
		}
		return nil, err
	}

	return &movie, nil
}

func (r *movieRepositoryImpl) FindByGenre(ctx context.Context, genreID int, limit, skip int) ([]models.Movie, error) {
	filter := bson.M{"genre.genre_id": genreID}
	opts := []*options.FindOptions{
		{
			Limit: &[]int64{int64(limit)}[0],
			Skip:  &[]int64{int64(skip)}[0],
			Sort:  bson.M{"ranking.ranking_value": -1},
		},
	}

	return r.FindAll(ctx, filter, opts)
}

func (r *movieRepositoryImpl) FindByGenres(ctx context.Context, genreIDs []int, limit int) ([]models.Movie, error) {
	filter := bson.M{"genre.genre_id": bson.M{"$in": genreIDs}}
	opts := []*options.FindOptions{
		{
			Limit: &[]int64{int64(limit)}[0],
			Sort:  bson.M{"ranking.ranking_value": -1},
		},
	}

	return r.FindAll(ctx, filter, opts)
}

func (r *movieRepositoryImpl) Update(ctx context.Context, id string, update bson.M) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrMovieNotFound
	}

	return nil
}

func (r *movieRepositoryImpl) Delete(ctx context.Context, id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrMovieNotFound
	}

	return nil
}

func (r *movieRepositoryImpl) Count(ctx context.Context, filter bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}

func (r *movieRepositoryImpl) MovieExists(ctx context.Context, imdbID string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"imdb_id": imdbID})
	return count > 0, err
}
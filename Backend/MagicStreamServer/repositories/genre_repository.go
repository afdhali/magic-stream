package repositories

import (
	"context"
	"errors"

	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrGenreNotFound = errors.New("genre not found")
)

// GenreRepository defines the interface for genre data operations
type GenreRepository interface {
	Create(ctx context.Context, genre *models.Genre) error
	FindAll(ctx context.Context) ([]models.Genre, error)
	FindByID(ctx context.Context, genreID int) (*models.Genre, error)
	FindByIDs(ctx context.Context, genreIDs []int) ([]models.Genre, error)
	ValidateGenres(ctx context.Context, genres []models.Genre) (bool, error)
	SeedGenres(ctx context.Context, genres []models.Genre) error
	Count(ctx context.Context) (int64, error)
}

// genreRepositoryImpl implements GenreRepository
type genreRepositoryImpl struct {
	collection *mongo.Collection
}

// NewGenreRepository creates a new genre repository
func NewGenreRepository(collection *mongo.Collection) GenreRepository {
	return &genreRepositoryImpl{
		collection: collection,
	}
}

func (r *genreRepositoryImpl) Create(ctx context.Context, genre *models.Genre) error {
	_, err := r.collection.InsertOne(ctx, genre)
	return err
}

func (r *genreRepositoryImpl) FindAll(ctx context.Context) ([]models.Genre, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var genres []models.Genre
	if err := cursor.All(ctx, &genres); err != nil {
		return nil, err
	}

	return genres, nil
}

func (r *genreRepositoryImpl) FindByID(ctx context.Context, genreID int) (*models.Genre, error) {
	var genre models.Genre
	err := r.collection.FindOne(ctx, bson.M{"genre_id": genreID}).Decode(&genre)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrGenreNotFound
		}
		return nil, err
	}

	return &genre, nil
}

func (r *genreRepositoryImpl) FindByIDs(ctx context.Context, genreIDs []int) ([]models.Genre, error) {
	filter := bson.M{"genre_id": bson.M{"$in": genreIDs}}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var genres []models.Genre
	if err := cursor.All(ctx, &genres); err != nil {
		return nil, err
	}

	return genres, nil
}

func (r *genreRepositoryImpl) ValidateGenres(ctx context.Context, genres []models.Genre) (bool, error) {
	if len(genres) == 0 {
		return false, nil
	}

	genreIDs := make([]int, len(genres))
	for i, genre := range genres {
		genreIDs[i] = genre.GenreID
	}

	count, err := r.collection.CountDocuments(ctx, bson.M{"genre_id": bson.M{"$in": genreIDs}})
	if err != nil {
		return false, err
	}

	return int(count) == len(genreIDs), nil
}

func (r *genreRepositoryImpl) SeedGenres(ctx context.Context, genres []models.Genre) error {
	// Check if genres already exist
	count, err := r.Count(ctx)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // Already seeded
	}

	var documents []interface{}
	for _, genre := range genres {
		documents = append(documents, genre)
	}

	_, err = r.collection.InsertMany(ctx, documents)
	return err
}

func (r *genreRepositoryImpl) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}
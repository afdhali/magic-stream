package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, userID string) (*models.User, error)
	UpdateFavoriteGenres(ctx context.Context, userID string, genres []models.Genre) error
	UserExists(ctx context.Context, email string) (bool, error)
}

// userRepositoryImpl implements UserRepository
type userRepositoryImpl struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new user repository
func NewUserRepository(collection *mongo.Collection) UserRepository {
	return &userRepositoryImpl{
		collection: collection,
	}
}

func (r *userRepositoryImpl) Create(ctx context.Context, user *models.User) error {
	// Check if user already exists
	exists, err := r.UserExists(ctx, user.Email)
	if err != nil {
		return err
	}
	if exists {
		return ErrUserAlreadyExists
	}

	_, err = r.collection.InsertOne(ctx, user)
	return err
}

func (r *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) FindByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) UpdateFavoriteGenres(ctx context.Context, userID string, genres []models.Genre) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"favourite_genres": genres,
			"updated_at":       time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *userRepositoryImpl) UserExists(ctx context.Context, email string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"email": email})
	return count > 0, err
}
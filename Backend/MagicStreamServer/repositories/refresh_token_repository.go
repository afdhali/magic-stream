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
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
)

// RefreshTokenRepository defines the interface for refresh token data operations
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	FindByToken(ctx context.Context, token string, userID string) (*models.RefreshToken, error)
	RevokeUserTokens(ctx context.Context, userID string) error
	RevokeToken(ctx context.Context, tokenID string) error
	CleanupExpired(ctx context.Context) error
}

// refreshTokenRepositoryImpl implements RefreshTokenRepository
type refreshTokenRepositoryImpl struct {
	collection *mongo.Collection
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(collection *mongo.Collection) RefreshTokenRepository {
	return &refreshTokenRepositoryImpl{
		collection: collection,
	}
}

func (r *refreshTokenRepositoryImpl) Create(ctx context.Context, token *models.RefreshToken) error {
	_, err := r.collection.InsertOne(ctx, token)
	return err
}

func (r *refreshTokenRepositoryImpl) FindByToken(ctx context.Context, token string, userID string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := r.collection.FindOne(ctx, bson.M{
		"token":   token,
		"user_id": userID,
	}).Decode(&refreshToken)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, err
	}

	return &refreshToken, nil
}

func (r *refreshTokenRepositoryImpl) RevokeUserTokens(ctx context.Context, userID string) error {
	filter := bson.M{"user_id": userID, "revoked": false}
	update := bson.M{"$set": bson.M{"revoked": true, "updated_at": time.Now()}}

	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}

func (r *refreshTokenRepositoryImpl) RevokeToken(ctx context.Context, tokenID string) error {
	objectID, err := bson.ObjectIDFromHex(tokenID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"revoked": true, "updated_at": time.Now()}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrRefreshTokenNotFound
	}

	return nil
}

func (r *refreshTokenRepositoryImpl) CleanupExpired(ctx context.Context) error {
	filter := bson.M{"expires_at": bson.M{"$lt": time.Now()}}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}
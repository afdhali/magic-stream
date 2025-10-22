package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshToken struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    string    `bson:"user_id"`
	Token     string    `bson:"token"`
	ExpiresAt time.Time `bson:"expires_at"`
	CreatedAt time.Time `bson:"created_at"`
	Revoked   bool      `bson:"revoked"`
}
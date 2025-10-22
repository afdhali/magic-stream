package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// User is the MongoDB document model
type User struct {
	ID              bson.ObjectID `bson:"_id,omitempty"`
	UserID          string        `bson:"user_id"`
	FirstName       string        `bson:"first_name"`
	LastName        string        `bson:"last_name"`
	Email           string        `bson:"email"`
	Password        string        `bson:"password"` // hashed
	Role            string        `bson:"role"`
	CreatedAt       time.Time     `bson:"created_at"`
	UpdatedAt       time.Time     `bson:"updated_at"`
	FavouriteGenres []Genre       `bson:"favourite_genres"`
}

// UserRegister is used for incoming registration requests
type UserRegister struct {
	FirstName       string  `json:"first_name" binding:"required,min=2,max=100" example:"John"`
	LastName        string  `json:"last_name" binding:"required,min=2,max=100" example:"Doe"`
	Email           string  `json:"email" binding:"required,email" example:"john.doe@example.com"`
	Password        string  `json:"password" binding:"required,min=6" example:"password123"`
	FavouriteGenres []Genre `json:"favourite_genres" binding:"required,dive"`
}

// UserLogin is used for login requests
type UserLogin struct {
	Email    string `json:"email" binding:"required,email" example:"john.doe@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// UserResponse is the safe output model (no password, no internal IDs)
type UserResponse struct {
	UserID          string  `json:"user_id" example:"507f1f77bcf86cd799439011"`
	FirstName       string  `json:"first_name" example:"John"`
	LastName        string  `json:"last_name" example:"Doe"`
	Email           string  `json:"email" example:"john.doe@example.com"`
	Role            string  `json:"role" example:"USER"`
	Token           string  `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken    string  `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	FavouriteGenres []Genre `json:"favourite_genres"`
}
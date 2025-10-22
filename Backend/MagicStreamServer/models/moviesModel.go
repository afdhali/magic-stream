package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Genre represents a movie genre
type Genre struct {
	GenreID   int    `bson:"genre_id" json:"genre_id" binding:"required" example:"1"`
	GenreName string `bson:"genre_name" json:"genre_name" binding:"required,min=2,max=100" example:"Action"`
}

// Ranking represents movie ranking information
type Ranking struct {
	RankingValue int    `bson:"ranking_value" json:"ranking_value" binding:"required,min=1,max=10" example:"9"`
	RankingName  string `bson:"ranking_name" json:"ranking_name" binding:"required,min=2,max=50" example:"Masterpiece"`
}

// Movie represents a movie document in the database
type Movie struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty" example:"507f1f77bcf86cd799439011"`
	ImdbID      string        `bson:"imdb_id" json:"imdb_id" binding:"required,min=9,max=10" example:"tt0111161"`
	Title       string        `bson:"title" json:"title" binding:"required,min=2,max=500" example:"The Shawshank Redemption"`
	PosterPath  string        `bson:"poster_path" json:"poster_path" binding:"required,url" example:"https://image.tmdb.org/t/p/w500/q6y0Go1tsGEsmtFryDOJo3dEmqu.jpg"`
	YouTubeID   string        `bson:"youtube_id" json:"youtube_id" binding:"required,min=11,max=11" example:"6hB3S9bIaco"`
	Genre       []Genre       `bson:"genre" json:"genre" binding:"required,min=1,dive"`
	AdminReview string        `bson:"admin_review" json:"admin_review" binding:"omitempty,max=1000" example:"One of the greatest movies of all time"`
	Ranking     Ranking       `bson:"ranking" json:"ranking" binding:"required"`
}

// MovieCreateRequest for creating a new movie (without ID)
type MovieCreateRequest struct {
	ImdbID      string   `json:"imdb_id" binding:"required,min=9,max=10" example:"tt0111161"`
	Title       string   `json:"title" binding:"required,min=2,max=500" example:"The Shawshank Redemption"`
	PosterPath  string   `json:"poster_path" binding:"required,url" example:"https://image.tmdb.org/t/p/w500/q6y0Go1tsGEsmtFryDOJo3dEmqu.jpg"`
	YouTubeID   string   `json:"youtube_id" binding:"required,min=11,max=11" example:"6hB3S9bIaco"`
	Genre       []Genre  `json:"genre" binding:"required,min=1,dive"`
	AdminReview string   `json:"admin_review" binding:"omitempty,max=1000" example:"One of the greatest movies of all time"`
	Ranking     Ranking  `json:"ranking" binding:"required"`
}

// MovieUpdateRequest for updating an existing movie
type MovieUpdateRequest struct {
	ImdbID      string   `json:"imdb_id" binding:"omitempty,min=9,max=10" example:"tt0111161"`
	Title       string   `json:"title" binding:"omitempty,min=2,max=500" example:"The Shawshank Redemption"`
	PosterPath  string   `json:"poster_path" binding:"omitempty,url" example:"https://image.tmdb.org/t/p/w500/q6y0Go1tsGEsmtFryDOJo3dEmqu.jpg"`
	YouTubeID   string   `json:"youtube_id" binding:"omitempty,min=11,max=11" example:"6hB3S9bIaco"`
	Genre       []Genre  `json:"genre" binding:"omitempty,min=1,dive"`
	AdminReview string   `json:"admin_review" binding:"omitempty,max=1000" example:"Updated review"`
	Ranking     Ranking  `json:"ranking" binding:"omitempty"`
}

// MovieFilterParams for query parameters
type MovieFilterParams struct {
	Genre   string `form:"genre" example:"Action"`
	Ranking int    `form:"ranking" example:"7"`
	Limit   int    `form:"limit" example:"10"`
	Skip    int    `form:"skip" example:"0"`
}

// ToMovie converts MovieCreateRequest to Movie
func (req *MovieCreateRequest) ToMovie() Movie {
	return Movie{
		ID:          bson.NewObjectID(),
		ImdbID:      req.ImdbID,
		Title:       req.Title,
		PosterPath:  req.PosterPath,
		YouTubeID:   req.YouTubeID,
		Genre:       req.Genre,
		AdminReview: req.AdminReview,
		Ranking:     req.Ranking,
	}
}

// ToMap converts MovieUpdateRequest to map for MongoDB update
func (req *MovieUpdateRequest) ToMap() map[string]interface{} {
	update := make(map[string]interface{})
	
	if req.ImdbID != "" {
		update["imdb_id"] = req.ImdbID
	}
	if req.Title != "" {
		update["title"] = req.Title
	}
	if req.PosterPath != "" {
		update["poster_path"] = req.PosterPath
	}
	if req.YouTubeID != "" {
		update["youtube_id"] = req.YouTubeID
	}
	if len(req.Genre) > 0 {
		update["genre"] = req.Genre
	}
	if req.AdminReview != "" {
		update["admin_review"] = req.AdminReview
	}
	if req.Ranking.RankingValue > 0 {
		update["ranking"] = req.Ranking
	}
	
	return update
}
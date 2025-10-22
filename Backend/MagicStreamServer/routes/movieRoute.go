package routes

import (
	"context"
	"net/http"
	"strconv"
	"time"

	authservice "github.com/afdhali/magic-stream/Backend/MagicStreamServer/controllers/auth"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/database"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/middleware"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/models"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/repositories"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MovieHandler handles movie-related requests
type MovieHandler struct {
	tokenService *authservice.TokenService
	movieRepo    repositories.MovieRepository
	genreRepo    repositories.GenreRepository
}

// NewMovieHandler creates a new movie handler with dependencies injected
func NewMovieHandler(ts *authservice.TokenService, movieRepo repositories.MovieRepository, genreRepo repositories.GenreRepository) *MovieHandler {
	return &MovieHandler{
		tokenService: ts,
		movieRepo:    movieRepo,
		genreRepo:    genreRepo,
	}
}

// GetAll godoc
// @Summary      Get all movies
// @Description  Retrieve list of all movies with optional filtering and pagination
// @Tags         Movies
// @Produce      json
// @Param        genre query string false "Filter by genre name"
// @Param        ranking query int false "Filter by minimum ranking value"
// @Param        limit query int false "Limit results (default 10, max 100)"
// @Param        skip query int false "Skip results for pagination (default 0)"
// @Success      200 {object} MovieListResponse "List of movies with pagination info"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /movies [get]
func (h *MovieHandler) GetAll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Parse query parameters
	genreFilter := c.Query("genre")
	rankingFilter := c.Query("ranking")
	limitStr := c.DefaultQuery("limit", "10")
	skipStr := c.DefaultQuery("skip", "0")

	limit, _ := strconv.Atoi(limitStr)
	skip, _ := strconv.Atoi(skipStr)

	// Validate and adjust limits
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if skip < 0 {
		skip = 0
	}

	// Build filter
	filter := bson.M{}
	if genreFilter != "" {
		filter["genre.genre_name"] = bson.M{"$regex": genreFilter, "$options": "i"}
	}
	if rankingFilter != "" {
		rankingValue, _ := strconv.Atoi(rankingFilter)
		filter["ranking.ranking_value"] = bson.M{"$gte": rankingValue}
	}

	moviesColl := database.OpenCollection("movies")

	// Get total count for pagination
	totalCount, err := moviesColl.CountDocuments(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count movies"})
		return
	}

	// Find movies with pagination
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"ranking.ranking_value": -1}) // Sort by ranking descending

	cursor, err := moviesColl.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies"})
		return
	}
	defer cursor.Close(ctx)

	var movies []models.Movie
	if err := cursor.All(ctx, &movies); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse movies"})
		return
	}

	// Return with pagination info
	c.JSON(http.StatusOK, gin.H{
		"data": movies,
		"pagination": gin.H{
			"total":       totalCount,
			"limit":       limit,
			"skip":        skip,
			"total_pages": (totalCount + int64(limit) - 1) / int64(limit),
			"current_page": (skip / limit) + 1,
		},
	})
}

// GetByID godoc
// @Summary      Get movie by ID
// @Description  Retrieve a single movie by its MongoDB ObjectID or IMDb ID
// @Tags         Movies
// @Produce      json
// @Param        id path string true "Movie ID (ObjectID or IMDb ID)"
// @Success      200 {object} models.Movie "Movie details"
// @Failure      400 {object} ErrorResponse "Invalid ID format"
// @Failure      404 {object} ErrorResponse "Movie not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /movies/{id} [get]
func (h *MovieHandler) GetByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := c.Param("id")
	moviesColl := database.OpenCollection("movies")

	var movie models.Movie
	var err error

	// Try to find by ObjectID first
	objectID, parseErr := bson.ObjectIDFromHex(id)
	if parseErr == nil {
		// Valid ObjectID format, search by _id
		err = moviesColl.FindOne(ctx, bson.M{"_id": objectID}).Decode(&movie)
	} else {
		// Not a valid ObjectID, try IMDb ID
		err = moviesColl.FindOne(ctx, bson.M{"imdb_id": id}).Decode(&movie)
	}

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movie"})
		return
	}

	c.JSON(http.StatusOK, movie)
}

// GetByGenre godoc
// @Summary      Get movies by genre
// @Description  Retrieve movies filtered by specific genre
// @Tags         Movies
// @Produce      json
// @Param        genre_id path int true "Genre ID"
// @Param        limit query int false "Limit results (default 10)"
// @Param        skip query int false "Skip results (default 0)"
// @Success      200 {object} MovieListResponse "List of movies"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /movies/genre/{genre_id} [get]
func (h *MovieHandler) GetByGenre(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	genreIDStr := c.Param("genre_id")
	genreID, err := strconv.Atoi(genreIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid genre ID"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	skipStr := c.DefaultQuery("skip", "0")
	limit, _ := strconv.Atoi(limitStr)
	skip, _ := strconv.Atoi(skipStr)

	if limit <= 0 || limit > 100 {
		limit = 10
	}

	filter := bson.M{"genre.genre_id": genreID}
	moviesColl := database.OpenCollection("movies")

	totalCount, _ := moviesColl.CountDocuments(ctx, filter)

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"ranking.ranking_value": -1})

	cursor, err := moviesColl.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies"})
		return
	}
	defer cursor.Close(ctx)

	var movies []models.Movie
	if err := cursor.All(ctx, &movies); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": movies,
		"pagination": gin.H{
			"total": totalCount,
			"limit": limit,
			"skip":  skip,
		},
	})
}

// GetRecommendedForUser godoc
// @Summary      Get recommended movies for user
// @Description  Get movies based on user's favorite genres
// @Tags         Movies
// @Security     BearerAuth
// @Produce      json
// @Param        limit query int false "Limit results (default 20)"
// @Success      200 {array} models.Movie "Recommended movies"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /movies/recommendations [get]
func (h *MovieHandler) GetRecommendedForUser(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get user's favorite genres
	usersColl := database.OpenCollection("users")
	var user models.User
	err := usersColl.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user data"})
		return
	}

	if len(user.FavouriteGenres) == 0 {
		c.JSON(http.StatusOK, []models.Movie{})
		return
	}

	// Extract genre IDs
	genreIDs := make([]int, len(user.FavouriteGenres))
	for i, genre := range user.FavouriteGenres {
		genreIDs[i] = genre.GenreID
	}

	// Find movies matching user's favorite genres
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	filter := bson.M{"genre.genre_id": bson.M{"$in": genreIDs}}
	moviesColl := database.OpenCollection("movies")

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.M{"ranking.ranking_value": -1})

	cursor, err := moviesColl.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recommendations"})
		return
	}
	defer cursor.Close(ctx)

	var movies []models.Movie
	if err := cursor.All(ctx, &movies); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse movies"})
		return
	}

	c.JSON(http.StatusOK, movies)
}

// Create godoc
// @Summary      Create new movie
// @Description  Create a new movie (Admin only)
// @Tags         Movies
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        movie body models.MovieCreateRequest true "Movie data"
// @Success      201 {object} models.Movie "Movie created"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Admin access required"
// @Failure      409 {object} ErrorResponse "Movie already exists"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /movies [post]
func (h *MovieHandler) Create(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var req models.MovieCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate genres exist
	if valid, err := ValidateGenres(ctx, req.Genre); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate genres"})
		return
	} else if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "One or more genres are invalid"})
		return
	}

	moviesColl := database.OpenCollection("movies")

	// Check if movie with same IMDb ID already exists
	var existingMovie models.Movie
	err := moviesColl.FindOne(ctx, bson.M{"imdb_id": req.ImdbID}).Decode(&existingMovie)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Movie with this IMDb ID already exists"})
		return
	}

	// Convert request to movie
	movie := req.ToMovie()

	// Insert movie
	_, err = moviesColl.InsertOne(ctx, movie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create movie"})
		return
	}

	c.JSON(http.StatusCreated, movie)
}

// Update godoc
// @Summary      Update movie
// @Description  Update an existing movie (Admin only)
// @Tags         Movies
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Movie ID"
// @Param        movie body models.MovieUpdateRequest true "Updated movie data"
// @Success      200 {object} models.Movie "Movie updated"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Admin access required"
// @Failure      404 {object} ErrorResponse "Movie not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /movies/{id} [put]
func (h *MovieHandler) Update(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := c.Param("id")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID format"})
		return
	}

	var req models.MovieUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate genres if provided
	if len(req.Genre) > 0 {
		if valid, err := ValidateGenres(ctx, req.Genre); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate genres"})
			return
		} else if !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "One or more genres are invalid"})
			return
		}
	}

	moviesColl := database.OpenCollection("movies")

	// Update movie
	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": req.ToMap()}

	result, err := moviesColl.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update movie"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	// Get updated movie
	var updatedMovie models.Movie
	err = moviesColl.FindOne(ctx, filter).Decode(&updatedMovie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated movie"})
		return
	}

	c.JSON(http.StatusOK, updatedMovie)
}

// Delete godoc
// @Summary      Delete movie
// @Description  Delete a movie (Admin only)
// @Tags         Movies
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Movie ID"
// @Success      200 {object} MessageResponse "Movie deleted"
// @Failure      400 {object} ErrorResponse "Invalid ID format"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Admin access required"
// @Failure      404 {object} ErrorResponse "Movie not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /movies/{id} [delete]
func (h *MovieHandler) Delete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := c.Param("id")
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID format"})
		return
	}

	moviesColl := database.OpenCollection("movies")
	result, err := moviesColl.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete movie"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Movie deleted successfully"})
}

// MovieListResponse for Swagger documentation
type MovieListResponse struct {
	Data       []models.Movie `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
}

// PaginationInfo for Swagger documentation
type PaginationInfo struct {
	Total       int64 `json:"total"`
	Limit       int   `json:"limit"`
	Skip        int   `json:"skip"`
	TotalPages  int64 `json:"total_pages"`
	CurrentPage int   `json:"current_page"`
}
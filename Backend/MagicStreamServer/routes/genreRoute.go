package routes

import (
	"context"
	"net/http"
	"strconv"
	"time"

	authservice "github.com/afdhali/magic-stream/Backend/MagicStreamServer/controllers/auth"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/database"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/models"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/repositories"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/utils"
	"github.com/gin-gonic/gin"
)

// GenreHandler handles genre-related requests
type GenreHandler struct {
	tokenService *authservice.TokenService
	genreRepo    repositories.GenreRepository
}

// NewGenreHandler creates a new genre handler with dependencies injected
func NewGenreHandler(ts *authservice.TokenService, genreRepo repositories.GenreRepository) *GenreHandler {
	return &GenreHandler{
		tokenService: ts,
		genreRepo:    genreRepo,
	}
}

// GetAllGenres godoc
// @Summary      Get all available genres
// @Description  Retrieve list of all movie genres for user selection
// @Tags         Genres
// @Produce      json
// @Success      200 {array} models.Genre "List of genres"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /genres [get]
func (h *GenreHandler) GetAllGenres(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	genres, err := h.genreRepo.FindAll(ctx)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, genres)
}

// GetGenreByID godoc
// @Summary      Get genre by ID
// @Description  Retrieve a single genre by its ID
// @Tags         Genres
// @Produce      json
// @Param        id path int true "Genre ID"
// @Success      200 {object} models.Genre "Genre details"
// @Failure      404 {object} ErrorResponse "Genre not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /genres/{id} [get]
func (h *GenreHandler) GetGenreByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	genreIDStr := c.Param("id")
	genreID, err := strconv.Atoi(genreIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid genre ID"})
		return
	}

	genre, err := h.genreRepo.FindByID(ctx, genreID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, genre)
}

// SeedGenres godoc
// @Summary      Seed initial genres (Admin only)
// @Description  Populate database with initial genre data
// @Tags         Genres
// @Security     BearerAuth
// @Produce      json
// @Success      201 {object} MessageResponse "Genres seeded successfully"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Admin access required"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /genres/seed [post]
func (h *GenreHandler) SeedGenres(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Default genres
	defaultGenres := []models.Genre{
		{GenreID: 1, GenreName: "Comedy"},
		{GenreID: 2, GenreName: "Drama"},
		{GenreID: 3, GenreName: "Western"},
		{GenreID: 4, GenreName: "Fantasy"},
		{GenreID: 5, GenreName: "Thriller"},
		{GenreID: 6, GenreName: "Sci-Fi"},
		{GenreID: 7, GenreName: "Action"},
		{GenreID: 8, GenreName: "Mystery"},
		{GenreID: 9, GenreName: "Crime"},
	}

	err := h.genreRepo.SeedGenres(ctx, defaultGenres)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Genres seeded successfully",
		"count":   len(defaultGenres),
	})
}

// ValidateGenres checks if provided genre IDs exist in database
func ValidateGenres(ctx context.Context, genres []models.Genre) (bool, error) {
	if len(genres) == 0 {
		return false, nil
	}

	// Extract genre IDs
	genreIDs := make([]int, len(genres))
	for i, genre := range genres {
		genreIDs[i] = genre.GenreID
	}

	// Check if all genres exist
	_, err := GetGenresByIDs(ctx, genreIDs)
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetGenresByIDs retrieves full genre data by IDs
func GetGenresByIDs(ctx context.Context, genreIDs []int) ([]models.Genre, error) {
	genreRepo := repositories.NewGenreRepository(database.OpenCollection("genres"))
	return genreRepo.FindByIDs(ctx, genreIDs)
}
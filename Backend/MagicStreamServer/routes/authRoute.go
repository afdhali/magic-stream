package routes

import (
	"context"
	"errors"
	"net/http"
	"time"

	authservice "github.com/afdhali/magic-stream/Backend/MagicStreamServer/controllers/auth"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/middleware"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/models"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/repositories"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	tokenService *authservice.TokenService
	userRepo     repositories.UserRepository
	genreRepo    repositories.GenreRepository
}

// NewAuthHandler creates a new auth handler with dependencies injected
func NewAuthHandler(ts *authservice.TokenService, userRepo repositories.UserRepository, genreRepo repositories.GenreRepository) *AuthHandler {
	return &AuthHandler{
		tokenService: ts,
		userRepo:     userRepo,
		genreRepo:    genreRepo,
	}
}

// Error response structure for Swagger
type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

// Success message response
type MessageResponse struct {
	Message string `json:"message" example:"success message"`
}

// Token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// Token refresh response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// UpdateFavoriteGenresRequest for updating favorite genres
type UpdateFavoriteGenresRequest struct {
	FavouriteGenres []models.Genre `json:"favourite_genres" binding:"required,dive"`
}

// Register godoc
// @Summary      Register new user
// @Description  Create a new user account with email and password
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body models.UserRegister true "User registration data"
// @Success      201 {object} models.UserResponse "Successfully registered"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Failure      409 {object} ErrorResponse "User already exists"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserRegister
	if !utils.ValidateRequest(c, &req) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Validate genres
	if len(req.FavouriteGenres) > 0 {
		valid, err := h.genreRepo.ValidateGenres(ctx, req.FavouriteGenres)
		if err != nil {
			utils.HandleError(c, err)
			return
		}
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "One or more genres are invalid. Use /api/v1/genres to get valid genres",
			})
			return
		}
	}

	// Hash password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Create user
	userID := bson.NewObjectID().Hex()
	newUser := models.User{
		UserID:          userID,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Email:           req.Email,
		Password:        hashedPassword,
		Role:            "USER",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		FavouriteGenres: req.FavouriteGenres,
	}

	// Insert user
	err = h.userRepo.Create(ctx, &newUser)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	// Generate tokens
	tokenPair, err := h.tokenService.GenerateTokenPair(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Build response
	resp := buildUserResponse(newUser, tokenPair)
	c.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user with email and password
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        credentials body models.UserLogin true "Login credentials"
// @Success      200 {object} models.UserResponse "Successfully logged in"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Failure      401 {object} ErrorResponse "Invalid credentials"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLogin
	if !utils.ValidateRequest(c, &req) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find user
	user, err := h.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Verify password
	if !verifyPassword(user.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate tokens
	tokenPair, err := h.tokenService.GenerateTokenPair(user.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Build response
	resp := buildUserResponse(*user, tokenPair)
	c.JSON(http.StatusOK, resp)
}

// Logout godoc
// @Summary      User logout
// @Description  Revoke all refresh tokens for the authenticated user
// @Tags         Authentication
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} MessageResponse "Successfully logged out"
// @Failure      401 {object} ErrorResponse "User not authenticated"
// @Failure      500 {object} ErrorResponse "Failed to logout"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := h.tokenService.RevokeRefreshTokens(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Get a new access token using a valid refresh token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        token body RefreshTokenRequest true "Refresh token"
// @Success      200 {object} RefreshTokenResponse "New tokens issued"
// @Failure      400 {object} ErrorResponse "Refresh token is required"
// @Failure      401 {object} ErrorResponse "Invalid or expired refresh token"
// @Failure      500 {object} ErrorResponse "Token refresh failed"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token is required"})
		return
	}

	tokenPair, err := h.tokenService.UseRefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, authservice.ErrInvalidToken) || errors.Is(err, authservice.ErrRevokedToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token refresh failed"})
		return
	}

	c.JSON(http.StatusOK, RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Get authenticated user's profile information
// @Tags         Authentication
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} models.User "User profile"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "User not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /auth/me [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	// Don't send password
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

// UpdateFavoriteGenres godoc
// @Summary      Update user's favorite genres
// @Description  Update authenticated user's favorite genres
// @Tags         Authentication
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        genres body UpdateFavoriteGenresRequest true "Favorite genres"
// @Success      200 {object} models.User "Updated user profile"
// @Failure      400 {object} ErrorResponse "Invalid request or genres"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /auth/favorite-genres [put]
func (h *AuthHandler) UpdateFavoriteGenres(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req UpdateFavoriteGenresRequest
	if !utils.ValidateRequest(c, &req) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Validate genres
	if len(req.FavouriteGenres) > 0 {
		valid, err := h.genreRepo.ValidateGenres(ctx, req.FavouriteGenres)
		if err != nil {
			utils.HandleError(c, err)
			return
		}
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "One or more genres are invalid. Use /api/v1/genres to get valid genres",
			})
			return
		}
	}

	// Update user's favorite genres
	err := h.userRepo.UpdateFavoriteGenres(ctx, userID, req.FavouriteGenres)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	// Get updated user
	user, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, user)
}

// ============================================================================
// Helper Functions
// ============================================================================

// hashPassword hashes a plain text password
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyPassword compares hashed password with plain text password
func verifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// buildUserResponse constructs a UserResponse from User and TokenPair
func buildUserResponse(user models.User, tokens *models.TokenPair) models.UserResponse {
	return models.UserResponse{
		UserID:          user.UserID,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Email:           user.Email,
		Role:            user.Role,
		Token:           tokens.AccessToken,
		RefreshToken:    tokens.RefreshToken,
		FavouriteGenres: user.FavouriteGenres,
	}
}
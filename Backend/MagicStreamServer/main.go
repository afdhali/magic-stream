package main

import (
	"fmt"

	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/config"
	authservice "github.com/afdhali/magic-stream/Backend/MagicStreamServer/controllers/auth"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/database"
	_ "github.com/afdhali/magic-stream/Backend/MagicStreamServer/docs"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/middleware"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/repositories"
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/routes"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Magic Stream API
// @version         1.0
// @description     API Server untuk Magic Stream - Streaming platform with authentication

// @contact.name   API Support
// @contact.email  support@magicstream.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:5000
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @schemes http https
func main() {
	// Load configuration
	cfg := config.LoadConfig()
	gin.SetMode(cfg.GinMode)

	// Database connection
	database.Connect()
	defer database.Disconnect()

	// Initialize router
	router := gin.Default()

	// Global middlewares
	router.Use(middleware.SecureHeaders())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Initialize database connection
	database.Connect()
	defer database.Disconnect()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(database.OpenCollection("users"))
	movieRepo := repositories.NewMovieRepository(database.OpenCollection("movies"))
	genreRepo := repositories.NewGenreRepository(database.OpenCollection("genres"))
	refreshTokenRepo := repositories.NewRefreshTokenRepository(database.OpenCollection("refresh_token"))

	// Initialize services
	tokenService := authservice.NewTokenService(cfg, refreshTokenRepo)

	// Setup routes
	setupRoutes(router, tokenService, userRepo, movieRepo, genreRepo)

	// Start server
	fmt.Printf("🚀 Server running on http://localhost:%s\n", cfg.Port)
	fmt.Printf("📚 Swagger UI: http://localhost:%s/swagger/index.html\n", cfg.Port)
	fmt.Printf("📋 Health Check: http://localhost:%s/api/v1/health\n", cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}

// setupRoutes configures all application routes
func setupRoutes(router *gin.Engine, ts *authservice.TokenService, userRepo repositories.UserRepository, movieRepo repositories.MovieRepository, genreRepo repositories.GenreRepository) {
	// API v1 group
	v1 := router.Group("/api/v1")

	// Health check endpoint
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": "1.0",
			"message": "Magic Stream API is running",
		})
	})

	// Feature routes
	setupAuthRoutes(v1, ts, userRepo, genreRepo)
	setupGenreRoutes(v1, ts, genreRepo)
	setupMovieRoutes(v1, ts, movieRepo, genreRepo)
}

// setupAuthRoutes configures authentication related routes
func setupAuthRoutes(rg *gin.RouterGroup, ts *authservice.TokenService, userRepo repositories.UserRepository, genreRepo repositories.GenreRepository) {
	auth := rg.Group("/auth")

	// Initialize auth handler with token service
	authHandler := routes.NewAuthHandler(ts, userRepo, genreRepo)

	// Public routes (no authentication required)
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.RefreshToken)

	// Protected routes (authentication required)
	auth.POST("/logout", middleware.AuthMiddleware(ts), authHandler.Logout)
	auth.GET("/me", middleware.AuthMiddleware(ts), authHandler.GetProfile)
	auth.PUT("/favorite-genres", middleware.AuthMiddleware(ts), authHandler.UpdateFavoriteGenres)
}

// setupGenreRoutes configures genre related routes
func setupGenreRoutes(rg *gin.RouterGroup, ts *authservice.TokenService, genreRepo repositories.GenreRepository) {
	genres := rg.Group("/genres")

	genreHandler := routes.NewGenreHandler(ts, genreRepo)

	// Public routes
	genres.GET("", genreHandler.GetAllGenres)
	genres.GET("/:id", genreHandler.GetGenreByID)

	// Protected routes (admin only)
	genres.POST("/seed",
		middleware.AuthMiddleware(ts),
		middleware.AdminOnly(),
		genreHandler.SeedGenres,
	)
}

// setupMovieRoutes configures movie related routes
func setupMovieRoutes(rg *gin.RouterGroup, ts *authservice.TokenService, movieRepo repositories.MovieRepository, genreRepo repositories.GenreRepository) {
	movies := rg.Group("/movies")

	movieHandler := routes.NewMovieHandler(ts, movieRepo, genreRepo)

	// Public routes
	movies.GET("", movieHandler.GetAll)
	movies.GET("/:id", movieHandler.GetByID)
	movies.GET("/genre/:genre_id", movieHandler.GetByGenre)

	// Protected routes (user must be authenticated)
	movies.GET("/recommendations",
		middleware.AuthMiddleware(ts),
		movieHandler.GetRecommendedForUser,
	)

	// Admin only routes
	movies.POST("",
		middleware.AuthMiddleware(ts),
		middleware.AdminOnly(),
		movieHandler.Create,
	)
	movies.PUT("/:id",
		middleware.AuthMiddleware(ts),
		middleware.AdminOnly(),
		movieHandler.Update,
	)
	movies.DELETE("/:id",
		middleware.AuthMiddleware(ts),
		middleware.AdminOnly(),
		movieHandler.Delete,
	)
}
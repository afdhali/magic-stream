package middleware

import (
	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/config"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SwaggerMiddleware setup swagger documentation
// Ini opsional, bisa langsung di main.go juga
func SwaggerMiddleware() gin.HandlerFunc {
	return ginSwagger.WrapHandler(swaggerFiles.Handler)
}

// Kalau mau custom config swagger
func SwaggerWithConfig() gin.HandlerFunc {
	cfg := config.LoadConfig()
	url := ginSwagger.URL(cfg.BackendServerURI+"/swagger/doc.json") // The url pointing to API definition
	return ginSwagger.WrapHandler(swaggerFiles.Handler, url)
}
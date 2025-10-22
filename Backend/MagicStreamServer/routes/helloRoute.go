package routes

import "github.com/gin-gonic/gin"

// HelloHandler godoc
// @Summary      Test API
// @Description  Test apakah API jalan
// @Tags         Test
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /hello [get]
func HelloHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello from Magic Stream API",
	})
}
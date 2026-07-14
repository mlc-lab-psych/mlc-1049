package routes

import (
	"net/http"
	"os"

	v1 "github.com/WeatherGod3218/mlc-project-template/internal/routes/api/v1"
	"github.com/gin-gonic/gin"
)

func GetHomepage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"ServerHost": os.Getenv("SERVER_HOST"),
	})
}

func SetRoutes(router *gin.RouterGroup) {
	router.GET("", GetHomepage)

	api := router.Group("/api")
	v1.Routes(api)
}

package v1

import (
	"github.com/WeatherGod3218/mlc-project-template/internal/routes/api/v1/data"
	"github.com/WeatherGod3218/mlc-project-template/internal/routes/api/v1/results"
	"github.com/gin-gonic/gin"
)

func Routes(r *gin.RouterGroup) {
	v1 := r.Group("/v1")
	data.Routes(v1)
	results.Routes(v1)
}

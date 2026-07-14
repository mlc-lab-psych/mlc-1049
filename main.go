package main

import (
	"context"
	"io/fs"
	"net/http"
	"os"

	"embed"
	"html/template"

	"github.com/WeatherGod3218/mlc-project-template/internal/airtable"
	"github.com/WeatherGod3218/mlc-project-template/internal/firebase"
	"github.com/WeatherGod3218/mlc-project-template/internal/logging"
	"github.com/WeatherGod3218/mlc-project-template/internal/redis"
	"github.com/WeatherGod3218/mlc-project-template/internal/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//go:embed templates/*
var embeddedFS embed.FS

//go:embed public/*
var staticFS embed.FS

func main() {
	err := redis.InitRedis()
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{"error": err, "module": "main", "method": "main"}).Warn("error launching redis! Proceeding without redis")
	}
	firebase.InitFirebase(context.Background())

	if err := airtable.InitalizeAirtables(); err != nil {
		logging.Logger.WithFields(logrus.Fields{"error": err, "module": "main", "method": "LoadAllAirtables"}).Fatal("error initalizing airtables!")
	}

	router := gin.Default()
	router.Use(cors.Default())

	tmpl := template.Must(template.ParseFS(embeddedFS, "templates/*"))
	router.SetHTMLTemplate(tmpl)

	staticSub, err := fs.Sub(staticFS, "public")
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{"error": err, "module": "main", "method": "main"}).Fatal("error embedding static files!")
	}
	router.StaticFS("/static", http.FS(staticSub))

	routes.SetRoutes(router.Group("/", redis.RedisRateLimiter(1, 100)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	router.Run(":" + port)
}

package main

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"time"

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
		logging.Logger.Warn("error launching redis! Proceeding without redis")
	}

	ready := make(chan error, 1)
	go func() {
		var firebaseErr, airtableErr error
		for i := 0; i < 5; i++ {
			firebaseErr = firebase.InitFirebase()
			if firebaseErr == nil {
				break
			}
			logging.Logger.WithFields(logrus.Fields{"attempt": i, "error": firebaseErr}).Warn("firebase init failed, retrying")
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
		}

		for i := 0; i < 5; i++ {
			airtableErr = airtable.InitalizeAirtables()
			if airtableErr == nil {
				break
			}
			logging.Logger.WithFields(logrus.Fields{"attempt": i, "error": airtableErr}).Warn("airtable init failed, retrying")
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
		}
		ready <- errors.Join(firebaseErr, airtableErr)
	}()

	if err := <-ready; err != nil {
		logging.Logger.WithFields(logrus.Fields{"error": err}).Error("critical service init failed after retries")
		os.Exit(1)
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

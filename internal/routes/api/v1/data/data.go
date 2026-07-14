package data

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/WeatherGod3218/mlc-project-template/internal/airtable"
	"github.com/WeatherGod3218/mlc-project-template/internal/firebase"
	"github.com/WeatherGod3218/mlc-project-template/internal/models"

	"github.com/WeatherGod3218/mlc-project-template/internal/logging"
	"github.com/sirupsen/logrus"
)

// func replaceNullWithString(obj []map[string]interface{}) {
// 	for key, value := range obj {
// 		if value == nil {
// 			obj[key] = "null"
// 		} else if nested, ok := value.(map[string]interface{}); ok {
// 			replaceNullWithString(nested)
// 		}
// 	}
// }

func GetDataFromUser(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		logging.Logger.Warn("unable to load id")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to load prolific id",
		})
		return
	}

	if user.WorkerID == "" {
		logging.Logger.Warn("id was empty!")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "prolific ID was empty!",
		})
		return
	}

	logging.Logger.Info(user.WorkerID)
	tableName, session, err := firebase.GetUserData(user.WorkerID)
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{"error": err, "module": "api", "method": "GetData"}).Warn("error fetching airtable!")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id",
		})
		return
	}
	data := airtable.GetAirtableData(tableName)

	c.JSON(http.StatusOK, gin.H{
		"data":    data,
		"session": session,
	})
}

func SubmitResults(c *gin.Context) {
	var results []map[string]any

	err := c.ShouldBindJSON(&results)
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{"error": err, "module": "api", "method": "SubmitResults"}).Warn("error casting")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	// replaceNullWithString(results)

	err = firebase.PushResultsToDatabase(results)
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{"error": err, "module": "api", "method": "SubmitResults"}).Warn("error updating database!")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error saving results.", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully saved results!"})
}

func Routes(r *gin.RouterGroup) {
	data := r.Group("/data")

	data.PUT("/", GetDataFromUser)
}

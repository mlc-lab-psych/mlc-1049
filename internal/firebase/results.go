package firebase

import (
	"context"
	"fmt"
	"time"

	"github.com/WeatherGod3218/mlc-project-template/internal/logging"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func PushResultsToDatabase(cleanedData []map[string]any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	id, err := uuid.NewV7()
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{"error": err, "module": "firebase", "method": "PushToDatabase"}).Warn("error generating the UUID7")
	}
	err = database.NewRef(fmt.Sprintf("mlc1049/results/%s", id.String())).Set(ctx, cleanedData)

	return err
}

package firebase

import (
	"context"
	"fmt"
	"time"

	"firebase.google.com/go/v4/db"
	"github.com/WeatherGod3218/mlc-project-template/internal/airtable"
	"github.com/WeatherGod3218/mlc-project-template/internal/logging"
)

type UserData struct {
	AirtableList    []string `json:"airtable_list"`
	CurrentAirtable int      `json:"current_airtable"`
}

func GetUserData(userId string) (string, int8, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	keyLookup := fmt.Sprintf("mlc1049/userdata/%s", userId)

	var result UserData
	var session int8

	err := database.NewRef(keyLookup).Transaction(ctx, func(tn db.TransactionNode) (interface{}, error) {
		var raw any

		if err := tn.Unmarshal(&raw); err != nil {
			return nil, err
		}

		if raw == nil {
			list := airtable.GetRandomAirtableSet()
			result = UserData{
				AirtableList:    list,
				CurrentAirtable: 0,
			}
			session = 1

			next := 1
			if len(list) <= 1 {
				next = 0
			}

			return UserData{
				AirtableList:    list,
				CurrentAirtable: next,
			}, nil
		}

		var userData UserData
		if err := tn.Unmarshal(&userData); err != nil {
			return nil, err
		}

		result = userData
		session = int8(userData.CurrentAirtable) + 1

		newTable := userData.CurrentAirtable + 1
		if newTable >= len(userData.AirtableList) {
			newTable = 0
		}

		newUserData := UserData{
			AirtableList:    userData.AirtableList,
			CurrentAirtable: newTable,
		}

		return newUserData, nil
	})
	if err != nil {
		return "", 0, err
	}

	logging.Logger.Info(result.AirtableList[result.CurrentAirtable])
	logging.Logger.Info(session)

	return result.AirtableList[result.CurrentAirtable], session, nil
}

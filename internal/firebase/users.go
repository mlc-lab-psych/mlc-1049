package firebase

import (
	"context"
	"fmt"
	"time"

	"firebase.google.com/go/v4/db"
	"github.com/WeatherGod3218/mlc-project-template/internal/airtable"
)

type UserData struct {
	AirtableList    []string `json:"airtable_list"`
	CurrentAirtable int      `json:"current_airtable"`
}

func GetUserData(userId string) (string, int8, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
			result = UserData{
				AirtableList:    airtable.GetRandomAirtableSet(),
				CurrentAirtable: 0,
			}
			session = 1
			return result, nil
		}

		var userData UserData
		if err := tn.Unmarshal(&userData); err != nil {
			return nil, err
		}

		result = userData

		newTable := userData.CurrentAirtable + 1
		session = int8(newTable + 1)
		if newTable >= len(userData.AirtableList) {
			newTable = 0
			session = 1
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

	return result.AirtableList[result.CurrentAirtable], session, nil
}

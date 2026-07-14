package airtable

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"sync"

	"github.com/WeatherGod3218/mlc-project-template/internal/models"
)

func InitalizeAirtables() error {
	var airtableBases map[string]string

	AirtableSet = make(map[string]*models.AirtableQueue, 3)
	AirtableSetReverse = make(map[string]*models.AirtableQueue, 3)
	AirtableFullSet = make(map[string]*models.AirtableQueue, 6)

	if err := json.Unmarshal([]byte(os.Getenv("AIRTABLE_BASES")), &airtableBases); err != nil {
		return fmt.Errorf("error loading airtable bases %s", err)
	}

	for airtable, base := range airtableBases {
		var airtableTables map[string]string

		err := json.Unmarshal([]byte(os.Getenv(fmt.Sprintf("AIRTABLE_%s", airtable))), &airtableTables)
		if err != nil {
			return fmt.Errorf("error loading airtable table prefix%s", err)
		}

		reverse := strings.HasSuffix(airtable, "_REVERSE")

		index := 0

		airtableQueue := &models.AirtableQueue{
			Queue:   make([]string, 8),
			Mutex:   sync.Mutex{},
			Current: 0,
		}

		for tableName, tblKey := range airtableTables {
			combinedKey := fmt.Sprintf("%s%s", airtable, tableName)

			airtableQueue.Queue[index] = combinedKey
			if err := LoadAirtable(combinedKey, base, tblKey, reverse); err != nil {
				return fmt.Errorf("error loading airtable %s data %s", airtable, err)
			}
			index++
		}

		rand.Shuffle(len(airtableQueue.Queue), func(i, j int) {
			airtableQueue.Queue[i], airtableQueue.Queue[j] = airtableQueue.Queue[j], airtableQueue.Queue[i]
		})

		AirtableFullSet[airtable] = airtableQueue
		if reverse {
			AirtableSetReverse[airtable] = airtableQueue
		} else {
			AirtableSet[airtable] = airtableQueue
		}
	}

	return nil
}

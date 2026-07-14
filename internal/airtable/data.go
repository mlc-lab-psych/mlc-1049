package airtable

import (
	"maps"
	"math/rand/v2"
	"slices"
	"sync"
)

var mu sync.RWMutex
var currentAirtable int

func GetRandomAirtableSet() []string {
	randomBool := rand.N(2) == 1

	var newQueue []string
	if randomBool {
		newQueue = slices.Collect(maps.Keys(AirtableSet))
	} else {
		newQueue = slices.Collect(maps.Keys(AirtableSetReverse))
	}

	rand.Shuffle(len(newQueue), func(i, j int) {
		newQueue[i], newQueue[j] = newQueue[j], newQueue[i]
	})
	return newQueue
}

func GetAirtableData(airtable string) *SavedData {
	airtableData := AirtableFullSet[airtable]

	airtableData.Mutex.Lock()
	defer airtableData.Mutex.Unlock()

	fullKey := airtableData.Queue[airtableData.Current]
	airtableData.Current++

	if airtableData.Current >= len(airtableData.Queue) {
		airtableData.Current = 0
	}

	return LoadedAirtables[fullKey]
}

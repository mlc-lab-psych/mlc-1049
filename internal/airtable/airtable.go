package airtable

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"time"

	"net/http"
	"net/url"

	"github.com/WeatherGod3218/mlc-project-template/internal/logging"
	"github.com/WeatherGod3218/mlc-project-template/internal/models"
	"github.com/sirupsen/logrus"
)

type AirtableRecord struct {
	ID          string       `json:"id"`
	CreatedTime time.Time    `json:"createdTime"`
	Fields      AirtableData `json:"fields"`
}

type AirtableAPIResponse struct {
	Records []AirtableRecord `json:"records"`
	Offset  string           `json:"offset"`
}

type AirtableData struct {
	Block           int    `json:"block"`
	Item            string `json:"item"`
	Stimulus        string `json:"stimulus"`
	CorrectKey      string `json:"correct_key"`
	StimulusType    string `json:"stimulus_type"`
	Trial           int    `json:"trial"`
	Category        string `json:"category"`
	Order           int    `json:"order"`
	TrialType       string `json:"trial_type"`
	CategoryDisplay string `json:"category_display"`
	Association     string `json:"association"`
}

type SavedData struct {
	TestStimuli       []AirtableData     `json:"test_stimuli"`
	Images            []string           `json:"images"`
	CategoryDisplay   map[int][][]string `json:"category_display"`
	CategoryWordImage string             `json:"category_word_image"`
}

var LoadedAirtables map[string]*SavedData = make(map[string]*SavedData, 3*2*8)

var (
	AirtableSet        map[string]*models.AirtableQueue
	AirtableSetReverse map[string]*models.AirtableQueue
	AirtableFullSet    map[string]*models.AirtableQueue
)

func ProcessAirTable(airtable []AirtableRecord) (SavedData, error) {
	images := make([]string, 0)
	test_stimuli := make([]AirtableData, 0)
	categoryWordImage := "words_animate_Cat1.png"

	categoryDisplay := map[int][][]string{
		0:  make([][]string, 2),
		1:  make([][]string, 2),
		2:  make([][]string, 2),
		3:  make([][]string, 2),
		4:  make([][]string, 2),
		5:  make([][]string, 2),
		6:  make([][]string, 2),
		7:  make([][]string, 2),
		8:  make([][]string, 2),
		9:  make([][]string, 2),
		10: make([][]string, 2),
		11: make([][]string, 2),
		12: make([][]string, 2),
	}

	sort.Slice(airtable, func(i, j int) bool {
		return airtable[i].Fields.Trial < airtable[j].Fields.Trial
	})

	for _, record := range airtable {

		if record.Fields.Stimulus == "inert" && record.Fields.CorrectKey == "d" {
			categoryWordImage = "words_animate_Cat2.png"
		}

		if record.Fields.StimulusType == "image" {
			images = append(images, record.Fields.Stimulus)
		}

		onLeft := record.Fields.CorrectKey == "d"
		if onLeft {
			record.Fields.Association = "left"
		} else {
			record.Fields.Association = "right"
		}
		test_stimuli = append(test_stimuli, record.Fields)

		if _, ok := categoryDisplay[record.Fields.Block]; !ok {
			logging.Logger.Warnf("ProcessAirTable: unexpected Block value %d (record trial=%d, category_display=%q)", record.Fields.Block, record.Fields.Trial, record.Fields.CategoryDisplay)
			categoryDisplay[record.Fields.Block] = [][]string{
				make([]string, 0),
				make([]string, 0),
			}
		}

		idx := 1
		if onLeft {
			idx = 0
		}

		if !slices.Contains(categoryDisplay[record.Fields.Block][idx], record.Fields.CategoryDisplay) {
			categoryDisplay[record.Fields.Block][idx] = append(categoryDisplay[record.Fields.Block][idx], record.Fields.CategoryDisplay)
		}
	}

	for k := range categoryDisplay {
		sort.Slice(categoryDisplay[k][0], func(i, j int) bool {
			return len(categoryDisplay[k][0][i]) < len(categoryDisplay[k][0][j])
		})

		sort.Slice(categoryDisplay[k][1], func(i, j int) bool {
			return len(categoryDisplay[k][1][i]) < len(categoryDisplay[k][1][j])
		})
	}

	return SavedData{
		TestStimuli:       test_stimuli,
		Images:            images,
		CategoryDisplay:   categoryDisplay,
		CategoryWordImage: categoryWordImage,
	}, nil
}

func LoadAirtable(key string, base string, table string, reverse bool) error {
	rawTable, err := FetchAirtable(base, table)

	if err != nil {
		return err
	}

	newAirtable, err := ProcessAirTable(rawTable)

	if err != nil {
		return err
	}

	logging.Logger.WithFields(logrus.Fields{"module": "airtable", "method": "LoadAllAirtables"}).Info(fmt.Sprintf("Airtable Name: %s", key))

	LoadedAirtables[key] = &newAirtable
	return nil
}

func FetchAirtable(base string, table string) ([]AirtableRecord, error) {
	baseURL := fmt.Sprintf(
		"https://api.airtable.com/v0/%s/%s",
		url.PathEscape(base),
		url.PathEscape(table),
	)

	var allRecords []AirtableRecord
	var offset string

	client := &http.Client{}

	for {
		fetchURL := baseURL

		if offset != "" {
			fetchURL += "?offset=" + url.QueryEscape(offset)
		}

		req, err := http.NewRequest("GET", fetchURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("AIRTABLE_API_KEY")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf(
				"Airtable error: %d %s",
				resp.StatusCode,
				string(body),
			)
		}

		var result AirtableAPIResponse

		err = json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}

		allRecords = append(allRecords, result.Records...)

		if result.Offset == "" {
			break
		}

		offset = result.Offset
	}

	return allRecords, nil
}

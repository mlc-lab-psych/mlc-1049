package models

import "sync"

type AirtableQueue struct {
	Queue   []string   `json:"queue"`
	Mutex   sync.Mutex `json:"sync"`
	Current int        `json:"current"`
}

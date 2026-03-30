package pipeline

import "github.com/google/uuid"

// Blob - an index of a blob file
//
// Example: ["1", "2", "3"] || ["1.5", "6.2", "1.9"]
type Blob = string

type Job struct {
	JobID uuid.UUID
	Stage string
	Input []Blob
}

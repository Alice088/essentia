package workers

import (
	"github.com/google/uuid"
)

type Task struct {
	UUID      uuid.UUID
	ObjectKey string
}

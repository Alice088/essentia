package workers

import (
	"Alice088/essentia/pkg/s3"
)

type Job struct {
	Object  s3.Object
	Attempt int
}

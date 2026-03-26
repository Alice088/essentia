package s3

import (
	"fmt"

	"github.com/google/uuid"
)

type Object struct {
	Name uuid.UUID
	Ext  string
}

func NewPDF() Object {
	return Object{
		Ext:  "pdf",
		Name: uuid.New(),
	}
}

func (o Object) Key() string {
	return fmt.Sprintf("%s.%s", o.Name.String(), o.Ext)
}

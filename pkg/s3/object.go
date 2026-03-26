package s3

import (
	"fmt"

	"github.com/google/uuid"
)

type Object struct {
	Name uuid.UUID
	Ext  string
}

func MustToPDF(fromUuid string) Object {
	return Object{
		Ext:  "pdf",
		Name: uuid.MustParse(fromUuid),
	}
}

func NewPDF() Object {
	return Object{
		Ext:  "pdf",
		Name: uuid.New(),
	}
}

func ToTXT(object Object) Object {
	return Object{
		Ext:  "txt",
		Name: object.Name,
	}
}

func (o Object) Key() string {
	return fmt.Sprintf("%s.%s", o.Name.String(), o.Ext)
}

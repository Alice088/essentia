package process

import "os"

type CD struct {
	SCDO     SCDO    `json:"scdo"`
	Document os.File `json:"document"`
}

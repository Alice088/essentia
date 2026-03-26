package pdf_parser

import "Alice088/essentia/pkg/env"

type Parser struct {
	Config env.WorkersParsing
	TMP    TMP
}

func NewParser(config env.WorkersParsing) Parser {
	return Parser{
		Config: config,
	}
}

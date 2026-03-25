package pdf_reader

type ReadResponse struct {
	Text      string   `json:"text"`
	Error     string   `json:"error"`
	ErrorCode string   `json:"error_code,omitempty"`
	Metadata  Metadata `json:"metadata"`
}

type Metadata struct {
	Size int `json:"size"`
}

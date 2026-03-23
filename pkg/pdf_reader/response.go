package pdf_reader

type ReadResponse struct {
	Text     string   `json:"text"`
	Error    string   `json:"error"`
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Size int `json:"size"`
}

package process

type Chapter struct {
	Title          string    `json:"title"`
	CompressedText string    `json:"compressed_text"`
	LocalThesis    string    `json:"local_thesis"`
	KeyArguments   [5]string `json:"key_arguments"`
	KeyFacts       [5]string `json:"key_facts"`
}

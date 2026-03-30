package process

type Extract struct {
	LocalThesis    string    `json:"local_thesis"`
	KeyArguments   [5]string `json:"key_arguments"`
	KeyFacts       [5]string `json:"key_facts"`
	CompressedText string    `json:"compressed_text"`
}

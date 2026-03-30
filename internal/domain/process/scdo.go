package process

type SCDO struct {
	Title    string    `json:"title"`
	File     File      `json:"file"`
	Global   Global    `json:"global"`
	Chapters []Chapter `json:"chapters"`
}

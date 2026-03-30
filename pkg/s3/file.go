package s3

import "encoding/json"

type File struct {
	Name string
	Ext  string
	Path string
	JSON json.RawMessage
}

func NewJSON(name, path string, object any) File {
	bytes, _ := json.Marshal(object)

	return File{
		Name: name,
		Ext:  "json",
		Path: path,
		JSON: bytes,
	}
}

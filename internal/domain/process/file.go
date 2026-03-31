package process

import "Alice088/essentia/pkg/size"

type File struct {
	Name  string   `json:"name"`
	CSize size.KiB `json:"c_size_kb"`
	Size  size.KiB `json:"size_kb"`
}

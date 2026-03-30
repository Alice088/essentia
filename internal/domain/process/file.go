package process

import "Alice088/essentia/pkg/real_size"

type File struct {
	Name  string        `json:"name"`
	CSize real_size.KiB `json:"c_size_kb"`
	Size  real_size.KiB `json:"size_kb"`
}

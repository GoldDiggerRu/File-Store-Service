package model

// File represents a stored file in the domain model.
type File struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Path string `json:"path,omitempty"`
}

package storage

import (
	"io"

	"github.com/GoldDiggerRu/File-Store-Service/model"
)

// FileRepository defines storage operations for files.
type FileRepository interface {
	Save(name string, r io.Reader) (model.File, error)
	List() ([]model.File, error)
}

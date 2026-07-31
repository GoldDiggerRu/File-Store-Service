package storage

import (
	"io"
	"os"
	"path/filepath"

	"github.com/GoldDiggerRu/File-Store-Service/model"
)

// FileRepo is a filesystem-based implementation of FileRepository.
type FileRepo struct {
	Dir string
}

func NewFileRepo(dir string) (*FileRepo, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, err
		}
	}
	return &FileRepo{Dir: dir}, nil
}

func (r *FileRepo) Save(name string, reader io.Reader) (model.File, error) {
	p := filepath.Join(r.Dir, name)
	f, err := os.Create(p)
	if err != nil {
		return model.File{}, err
	}
	defer f.Close()
	n, err := io.Copy(f, reader)
	if err != nil {
		return model.File{}, err
	}
	return model.File{Name: name, Size: n, Path: p}, nil
}

func (r *FileRepo) List() ([]model.File, error) {
	entries, err := os.ReadDir(r.Dir)
	if err != nil {
		return nil, err
	}
	var files []model.File
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, model.File{
			Name: info.Name(),
			Size: info.Size(),
			Path: filepath.Join(r.Dir, info.Name()),
		})
	}
	return files, nil
}

package main

import (
	"log"

	"github.com/GoldDiggerRu/File-Store-Service/app"
	"github.com/GoldDiggerRu/File-Store-Service/storage"
)

const (
	port       = ":8080"
	storageDir = "file_storage"
)

func main() {
	repo, err := storage.NewFileRepo(storageDir)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(repo, port); err != nil {
		log.Fatal(err)
	}
}

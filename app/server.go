package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/GoldDiggerRu/File-Store-Service/storage"
)

const defaultPort = ":8080"

// Run starts the HTTP server using the provided FileRepository.
func Run(repo storage.FileRepository, port string) error {
	if port == "" {
		port = defaultPort
	}
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/upload", makeUploadHandler(repo))
	http.HandleFunc("/list", makeListHandler(repo))

	log.Printf("File Store Service starting on %s", port)
	log.Printf("Endpoints available:\n  GET  /health\n  POST /upload\n  GET  /list")

	return http.ListenAndServe(port, nil)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"File Store Service"}`)
}

func makeUploadHandler(repo storage.FileRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "File too large", http.StatusBadRequest)
			return
		}
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error retrieving file", http.StatusBadRequest)
			return
		}
		defer file.Close()
		saved, err := repo.Save(handler.Filename, file)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := map[string]string{"message": "File uploaded successfully", "filename": saved.Name}
		json.NewEncoder(w).Encode(resp)
	}
}

func makeListHandler(repo storage.FileRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		files, err := repo.List()
		if err != nil {
			http.Error(w, "Error reading storage", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"files": files})
	}
}

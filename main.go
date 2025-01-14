package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var config *Config

type Config struct {
	IncomingPath  string
	OutcomingPath string
}

type Response struct {
	Message string `json:"message"`
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func listFilesHandler(w http.ResponseWriter, r *http.Request) {

	config, error := loadConfig()
	if error != nil {
		fmt.Println(error.Error())
		http.Error(w, error.Error(), http.StatusInternalServerError)
		return
	}

	dir, err := os.Open(config.IncomingPath)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	files, err := dir.Readdir(0)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}

	json.NewEncoder(w).Encode(fileNames)
}

func downloadFileHandler(w http.ResponseWriter, r *http.Request) {
	config, error := loadConfig()
	if error != nil {
		fmt.Println(error.Error())
		http.Error(w, error.Error(), http.StatusInternalServerError)
		return
	}
	fileName := r.URL.Query().Get("file")
	filePath := filepath.Join(config.IncomingPath, fileName)

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, file)
}

func moveFileHandler(w http.ResponseWriter, r *http.Request) {
	config, error := loadConfig()
	if error != nil {
		fmt.Println(error.Error())
		http.Error(w, error.Error(), http.StatusInternalServerError)
		return
	}
	fileName := r.URL.Query().Get("file")
	srcPath := filepath.Join(config.IncomingPath, fileName)
	dstPath := filepath.Join(config.OutcomingPath, fileName)

	err := os.Rename(srcPath, dstPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := Response{Message: "File moved successfully"}
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/listFiles", listFilesHandler)
	http.HandleFunc("/downloadFile", downloadFileHandler)
	http.HandleFunc("/moveFile", moveFileHandler)
	http.ListenAndServe(":8080", nil)
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"./docs"
	httpSwagger "github.com/swaggo/http-swagger"
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

// @title File Management API
// @version 1.0
// @description This is a sample server for managing files.
// @host localhost:8080
// @BasePath /

// listFilesHandler godoc
// @Summary List files
// @Description Get a list of files in the incoming directory
// @Tags files
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} Response
// @Router /listFiles [get]
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

// downloadFileHandler godoc
// @Summary Download file
// @Description Download a file from the incoming directory
// @Tags files
// @Produce octet-stream
// @Param file query string true "File name"
// @Success 200
// @Failure 404 {object} Response
// @Router /downloadFile [get]
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

// moveFileHandler godoc
// @Summary Move file
// @Description Move a file from the incoming directory to the outgoing directory
// @Tags files
// @Produce json
// @Param file query string true "File name"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /moveFile [post]
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
	docs.SwaggerInfo.Title = "File Management API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Description = "This is a sample server for managing files."
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/"

	http.HandleFunc("/listFiles", listFilesHandler)
	http.HandleFunc("/downloadFile", downloadFileHandler)
	http.HandleFunc("/moveFile", moveFileHandler)
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	http.ListenAndServe(":8080", nil)
}

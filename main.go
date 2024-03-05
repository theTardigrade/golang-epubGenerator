package main

import (
	"encoding/json"
	"os"
)

const (
	epubInfoFileExt  = ".json"
	epubInfoFileName = "epub_info" + epubInfoFileExt
)

type epubInfo struct {
	CoverImagePath string `json:"cover_image_path"`
	ContentPath    string `json:"content_path"`
	ISBN           string `json:"isbn"`
	StylesPath     string `json:"styles_path"`
	Title          string `json:"title"`
}

func main() {
	fileContent, err := os.ReadFile(epubInfoFileName)
	if err != nil {
		panic(err)
	}

	var ei epubInfo

	err = json.Unmarshal(fileContent, &ei)
	if err != nil {
		panic(err)
	}

	if err := generate(&ei); err != nil {
		panic(err)
	}
}

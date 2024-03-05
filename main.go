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
	ISBN   string `json:"isbn"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Paths  struct {
		CoverImage string `json:"cover_image"`
		Styles     string `json:"styles"`
		Text       string `json:"text"`
	} `json:"paths"`
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

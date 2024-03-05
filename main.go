package main

import (
	"encoding/json"
	_ "image/png"
	"os"
)

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

	epubInfoInit(&ei)

	if err := generate(&ei); err != nil {
		panic(err)
	}
}

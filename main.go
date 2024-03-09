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

	if err = json.Unmarshal(fileContent, &ei); err != nil {
		panic(err)
	}

	if err = epubInfoOutputInit(&ei); err != nil {
		panic(err)
	}

	if err = generate(&ei); err != nil {
		panic(err)
	}
}

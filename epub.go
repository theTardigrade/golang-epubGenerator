package main

import (
	"image"
	"os"
)

const (
	epubInfoFileExt  = ".json"
	epubInfoFileName = "epub_info" + epubInfoFileExt
)

type epubInfo struct {
	ISBN                 string `json:"isbn"`
	Title                string `json:"title"`
	Author               string `json:"author"`
	IncludeCopyrightPage bool   `json:"include_copyright_page"`
	Paths                struct {
		CoverImage string `json:"cover_image"`
		Styles     string `json:"styles"`
		Text       string `json:"text"`
	} `json:"paths"`

	coverImage       image.Image
	coverImageFormat string
}

type epubInfoInitHandler = func(*epubInfo) error

var (
	epubInfoInitHandlerList = []epubInfoInitHandler{
		epubInfoInitCoverImage,
	}
)

func epubInfoInit(ei *epubInfo) {
	for _, handler := range epubInfoInitHandlerList {
		handler(ei)
	}
}

func epubInfoInitCoverImage(ei *epubInfo) (err error) {
	if ei.Paths.CoverImage == "" {
		return
	}

	f, err := os.Open(ei.Paths.CoverImage)
	if err != nil {
		return
	}
	defer f.Close()

	image, imageFormat, err := image.Decode(f)
	if err != nil {
		return
	}

	ei.coverImage = image
	ei.coverImageFormat = imageFormat

	return
}

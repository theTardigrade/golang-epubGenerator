package main

import (
	"bytes"
	"image"
	"os"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
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
	text             []byte
	textHeadings     []string
}

type epubInfoInitHandler = func(*epubInfo) error

var (
	epubInfoInitHandlerList = []epubInfoInitHandler{
		epubInfoInitCoverImage,
		epubInfoInitText,
		epubInfoInitTextHeadings,
	}
)

func epubInfoInit(ei *epubInfo) (err error) {
	for _, handler := range epubInfoInitHandlerList {
		if err = handler(ei); err != nil {
			return
		}
	}

	return
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

func epubInfoInitText(ei *epubInfo) (err error) {
	b, err := os.ReadFile(ei.Paths.Text)
	if err != nil {
		panic(err)
	}

	p := parser.New()

	document := p.Parse(b)
	renderer := html.NewRenderer(html.RendererOptions{
		Flags: html.CommonFlags,
	})

	b = markdown.Render(document, renderer)

	b, err = minifier.Bytes("text/xml", b)
	if err != nil {
		return
	}

	ei.text = b

	return
}

func epubInfoInitTextHeadings(ei *epubInfo) (err error) {
	r := bytes.NewReader(ei.text)

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return
	}

	doc.Find("h1").Each(func(i int, s *goquery.Selection) {
		ei.textHeadings = append(ei.textHeadings, s.Text())

		s.SetAttr("id", "epub_generator_text_heading_"+strconv.Itoa(len(ei.textHeadings)))
	})

	docString, err := doc.Find("body").Html()
	if err != nil {
		return
	}

	ei.text = []byte(docString)

	return
}

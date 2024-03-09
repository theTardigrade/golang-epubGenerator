package main

import (
	"bytes"
	"errors"
	"image"
	"os"
	"path/filepath"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/iancoleman/strcase"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	epubInfoFileExt  = ".json"
	epubInfoFileName = "epub_info" + epubInfoFileExt
)

type epubInfo struct {
	ISBN                     string   `json:"isbn"`
	Title                    string   `json:"title"`
	Author                   string   `json:"author"`
	EditionNumber            int      `json:"edition_number"`
	Files                    []string `json:"files"`
	IncludeContentsPage      bool     `json:"include_contents_page"`
	IncludeCopyrightPage     bool     `json:"include_copyright_page"`
	ShouldCapitalizeHeadings bool     `json:"should_capitalize_headings"`
	Paths                    struct {
		CoverImage string `json:"cover_image"`
		Styles     string `json:"styles"`
		Text       string `json:"text"`
	} `json:"paths"`

	output struct {
		coverImage       image.Image
		coverImageFormat string
		text             []byte
		textHeadings     []string
		title            string
	}
}

type epubInfoOutputInitHandler = func(*epubInfo) error

var (
	epubInfoOutputInitHandlerList = []epubInfoOutputInitHandler{
		epubInfoOutputInitCoverImage,
		epubInfoOutputInitText,
		epubInfoOutputInitTextHeadings,
		epubInfoOutputInitOutputTitle,
	}
)

func epubInfoOutputInit(ei *epubInfo) (err error) {
	for _, handler := range epubInfoOutputInitHandlerList {
		if err = handler(ei); err != nil {
			return
		}
	}

	return
}

func epubInfoOutputInitCoverImage(ei *epubInfo) (err error) {
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

	ei.output.coverImage = image
	ei.output.coverImageFormat = imageFormat

	return
}

func epubInfoOutputInitText(ei *epubInfo) (err error) {
	b, err := os.ReadFile(ei.Paths.Text)
	if err != nil {
		panic(err)
	}

	switch filepath.Ext(ei.Paths.Text) {
	case ".md":
		p := parser.New()

		document := p.Parse(b)
		renderer := html.NewRenderer(html.RendererOptions{
			Flags: html.CommonFlags,
		})

		b = markdown.Render(document, renderer)
	case ".html", ".xhtml":
		r := bytes.NewReader(b)

		doc, err := goquery.NewDocumentFromReader(r)
		if err != nil {
			return err
		}

		docString, err := doc.Find("body").Html()
		if err != nil {
			return err
		}

		ei.output.text = []byte(docString)
	default:
		return errors.New("unrecognized text file extension")
	}

	b, err = minifier.Bytes("text/xml", b)
	if err != nil {
		return
	}

	ei.output.text = b

	return
}

func epubInfoOutputInitTextHeadings(ei *epubInfo) (err error) {
	r := bytes.NewReader(ei.output.text)

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return
	}

	var caser cases.Caser

	if ei.ShouldCapitalizeHeadings {
		caser = cases.Title(language.English)
	}

	doc.Find("h1,h2,h3,h4,h5,h6").Each(func(i int, s *goquery.Selection) {
		heading := s.Text()

		if ei.ShouldCapitalizeHeadings {
			heading = caser.String(heading)

			s.SetText(heading)
		}

		if s.Is("h1") {
			ei.output.textHeadings = append(ei.output.textHeadings, heading)

			s.SetAttr("id", "epub_generator_text_heading_"+strconv.Itoa(len(ei.output.textHeadings)))
		}
	})

	docString, err := doc.Find("body").Html()
	if err != nil {
		return
	}

	ei.output.text = []byte(docString)

	return
}

func epubInfoOutputInitOutputTitle(ei *epubInfo) (err error) {
	ei.output.title = strcase.ToSnake(ei.Title)

	return
}

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
	ISBN                         string   `json:"isbn"`
	Title                        string   `json:"title"`
	Author                       string   `json:"author"`
	EditionNumber                int      `json:"edition_number"`
	Files                        []string `json:"files"`
	IncludeContentsPage          bool     `json:"include_contents_page"`
	IncludeCopyrightPage         bool     `json:"include_copyright_page"`
	ShouldCapitalizeMainHeadings bool     `json:"should_capitalize_main_headings"`
	Paths                        struct {
		CoverImage string `json:"cover_image"`
		Styles     string `json:"styles"`
		Text       string `json:"text"`
	} `json:"paths"`

	coverImage       image.Image
	coverImageFormat string
	text             []byte
	textHeadings     []string
	outputTitle      string
}

type epubInfoInitHandler = func(*epubInfo) error

var (
	epubInfoInitHandlerList = []epubInfoInitHandler{
		epubInfoInitCoverImage,
		epubInfoInitText,
		epubInfoInitTextHeadings,
		epubInfoInitOutputTitle,
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

	switch filepath.Ext(ei.Paths.Text) {
	case ".md":
		p := parser.New()

		document := p.Parse(b)
		renderer := html.NewRenderer(html.RendererOptions{
			Flags: html.CommonFlags,
		})

		b = markdown.Render(document, renderer)
	case ".html", ".xhtml":
		r := bytes.NewReader(ei.text)

		doc, err := goquery.NewDocumentFromReader(r)
		if err != nil {
			return err
		}

		docString, err := doc.Find("body").Html()
		if err != nil {
			return err
		}

		ei.text = []byte(docString)
	default:
		return errors.New("unrecognized text file extension")
	}

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

	var caser cases.Caser

	if ei.ShouldCapitalizeMainHeadings {
		caser = cases.Title(language.English)
	}

	doc.Find("h1").Each(func(i int, s *goquery.Selection) {
		heading := s.Text()

		if ei.ShouldCapitalizeMainHeadings {
			heading = caser.String(heading)
		}

		ei.textHeadings = append(ei.textHeadings, heading)

		s.SetAttr("id", "epub_generator_text_heading_"+strconv.Itoa(len(ei.textHeadings)))

		if ei.ShouldCapitalizeMainHeadings {
			s.SetText(heading)
		}
	})

	docString, err := doc.Find("body").Html()
	if err != nil {
		return
	}

	ei.text = []byte(docString)

	return
}

func epubInfoInitOutputTitle(ei *epubInfo) (err error) {
	ei.outputTitle = strcase.ToSnake(ei.Title)

	return
}

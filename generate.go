package main

import (
	"archive/zip"
	"bytes"
	"image/png"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type generateHandler func(*epubInfo, *zip.Writer) error

var (
	generateHandlerList = []generateHandler{
		generateMimetype,
		generateContainer,
		generateStyles,
		generateCoverImage,
		generateCoverPage,
		generateTitlePage,
		generateTextPage,
		generateOCF,
		generateNCX,
	}
)

func generate(ei *epubInfo) (err error) {
	err = func() (err error) {
		archiveFile, err := os.Create("output.zip")
		if err != nil {
			return
		}
		defer archiveFile.Close()

		archiveWriter := zip.NewWriter(archiveFile)
		defer archiveWriter.Close()

		for _, handler := range generateHandlerList {
			if err = handler(ei, archiveWriter); err != nil {
				return
			}
		}

		return
	}()
	if err != nil {
		return
	}

	if err = os.Rename("output.zip", "output.epub"); err != nil {
		return
	}

	return
}

func generateMimetype(ei *epubInfo, archiveWriter *zip.Writer) (err error) {
	w, err := archiveWriter.CreateHeader(&zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	})
	if err != nil {
		return
	}

	if _, err = io.WriteString(w, "application/epub+zip"); err != nil {
		return
	}

	return
}

func generateContainer(ei *epubInfo, archiveWriter *zip.Writer) (err error) {
	w, err := archiveWriter.Create("META-INF/container.xml")
	if err != nil {
		return
	}

	var contentBuilder strings.Builder

	contentBuilder.WriteString(`<?xml version="1.0" encoding="utf-8"?>`)
	contentBuilder.WriteString(`<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">`)
	contentBuilder.WriteString(`<rootfiles>`)
	contentBuilder.WriteString(`<rootfile full-path="content.opf" media-type="application/oebps-package+xml" />`)
	contentBuilder.WriteString(`</rootfiles>`)
	contentBuilder.WriteString(`</container>`)

	if _, err = io.WriteString(w, contentBuilder.String()); err != nil {
		return
	}

	return
}

func generateStyles(ei *epubInfo, archiveWriter *zip.Writer) (err error) {
	w, err := archiveWriter.Create("styles.css")
	if err != nil {
		return
	}

	if ei.Paths.Styles != "" {
		var b []byte

		b, err = os.ReadFile(ei.Paths.Styles)
		if err != nil {
			return
		}

		b, err = minifier.Bytes("text/css", b)
		if err != nil {
			return
		}

		if _, err = w.Write(b); err != nil {
			return
		}
	}

	return
}

func generateCoverImage(ei *epubInfo, archiveWriter *zip.Writer) (err error) {
	if ei.coverImage == nil {
		return
	}

	w, err := archiveWriter.Create("cover.png")
	if err != nil {
		return
	}

	err = png.Encode(w, ei.coverImage)
	if err != nil {
		return
	}

	return
}

func generateCoverPage(ei *epubInfo, archiveWriter *zip.Writer) (err error) {
	if ei.coverImage == nil {
		return
	}

	w, err := archiveWriter.Create("cover.xhtml")
	if err != nil {
		return
	}

	var headerBuilder, bodyBuilder bytes.Buffer
	coverImageBounds := ei.coverImage.Bounds()
	coverImageWidthString := strconv.Itoa(coverImageBounds.Dx())
	coverImageHeightString := strconv.Itoa(coverImageBounds.Dy())

	headerBuilder.WriteString(`<style type="text/css">`)
	headerBuilder.WriteString(`@page{padding:0pt;margin:0pt}`)
	headerBuilder.WriteString(`body{text-align:center;padding:0pt;margin:0pt;}`)
	headerBuilder.WriteString(`</style>`)

	bodyBuilder.WriteString(`<div>`)
	bodyBuilder.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" width="100%" height="100%" viewBox="0 0 ` + coverImageWidthString + ` ` + coverImageHeightString + `" preserveAspectRatio="none">`)
	bodyBuilder.WriteString(`<image width="` + coverImageWidthString + `" height="` + coverImageHeightString + `" xlink:href="cover.png" />`)
	bodyBuilder.WriteString(`</svg>`)
	bodyBuilder.WriteString(`</div>`)

	if _, err = io.WriteString(w, xhtmlHeader("Cover", headerBuilder.String())); err != nil {
		return
	}

	if _, err = w.Write(bodyBuilder.Bytes()); err != nil {
		return
	}

	if _, err = io.WriteString(w, xhtmlFooter()); err != nil {
		return
	}

	return
}

func generateTitlePage(ei *epubInfo, archiveWriter *zip.Writer) (err error) {
	w, err := archiveWriter.Create("title.xhtml")
	if err != nil {
		return
	}

	var builder bytes.Buffer

	builder.WriteString(`<div class="title_area">`)
	builder.WriteString(`<h1 class="title">` + ei.Title + "</h1>")

	if ei.Author != "" {
		builder.WriteString(`<h2 class="author">` + ei.Author + "</h1>")
	}

	builder.WriteString(`</div>`)

	if _, err = io.WriteString(w, xhtmlHeader("Title", "")); err != nil {
		return
	}

	if _, err = w.Write(builder.Bytes()); err != nil {
		return
	}

	if _, err = io.WriteString(w, xhtmlFooter()); err != nil {
		return
	}

	return
}

func generateTextPage(ei *epubInfo, archiveWriter *zip.Writer) (err error) {
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

	w, err := archiveWriter.Create("text.xhtml")
	if err != nil {
		return
	}

	if _, err = io.WriteString(w, xhtmlHeader("Text", "")); err != nil {
		return
	}

	if _, err = w.Write(b); err != nil {
		return
	}

	if _, err = io.WriteString(w, xhtmlFooter()); err != nil {
		return
	}

	return
}

func generateOCF(ei *epubInfo, archiveWriter *zip.Writer) (err error) {
	w, err := archiveWriter.Create("content.opf")
	if err != nil {
		return
	}

	var builder strings.Builder

	builder.WriteString(`<?xml version="1.0" encoding="utf-8"?>`)
	builder.WriteString(`<package xmlns="http://www.idpf.org/2007/opf" version="2.0" unique-identifier="unique-id">`)
	builder.WriteString(`<metadata xmlns:opf="http://www.idpf.org/2007/opf" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:calibre="http://calibre.kovidgoyal.net/2009/metadata">`)
	builder.WriteString(`<dc:language>en</dc:language>`)
	builder.WriteString(`<dc:title>` + ei.Title + `</dc:title>`)

	if ei.Author != "" {
		builder.WriteString(`<dc:creator>` + ei.Author + `</dc:creator>`)
	}

	builder.WriteString(`<dc:identifier id="unique-id">` + ei.ISBN + `</dc:identifier>`)
	builder.WriteString(`</metadata>`)
	builder.WriteString(`<manifest>`)
	builder.WriteString(`<item id="styles" href="styles.css" media-type="text/css" />`)

	if ei.coverImage != nil {
		builder.WriteString(`<item id="cover_image" href="cover.png" media-type="image/png" />`)
		builder.WriteString(`<item id="cover_page" href="cover.xhtml" media-type="application/xhtml+xml" />`)
	}

	builder.WriteString(`<item id="title_page" href="title.xhtml" media-type="application/xhtml+xml" />`)
	builder.WriteString(`<item id="text_page" href="text.xhtml" media-type="application/xhtml+xml" />`)
	builder.WriteString(`<item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml" />`)
	builder.WriteString(`</manifest>`)
	builder.WriteString(`<spine toc="ncx">`)

	if ei.coverImage != nil {
		builder.WriteString(`<itemref idref="cover_page" />`)
	}

	builder.WriteString(`<itemref idref="title_page" />`)
	builder.WriteString(`<itemref idref="text_page" />`)
	builder.WriteString(`</spine>`)

	if ei.coverImage != nil {
		builder.WriteString(`<guide>`)
		builder.WriteString(`<reference type="cover_page" href="cover.xhtml" title="Cover" />`)
		builder.WriteString(`</guide>`)
	}

	builder.WriteString(`</package>`)

	if _, err = io.WriteString(w, builder.String()); err != nil {
		return
	}

	return
}

func generateNCX(ei *epubInfo, archiveWriter *zip.Writer) (err error) {
	w, err := archiveWriter.Create("toc.ncx")
	if err != nil {
		return
	}

	var contentBuilder strings.Builder
	var playOrder int

	contentBuilder.WriteString(`<?xml version="1.0" encoding="utf-8"?>`)
	contentBuilder.WriteString(`<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1" xml:lang="eng">`)
	contentBuilder.WriteString(`<head>`)
	contentBuilder.WriteString(`<meta name="dtb:uid" content="` + ei.ISBN + `" />`)
	contentBuilder.WriteString(`<meta name="dtb:totalPageCount" content="0" />`)
	contentBuilder.WriteString(`<meta name="dtb:maxPageNumber" content="0" />`)
	contentBuilder.WriteString(`</head>`)
	contentBuilder.WriteString(`<docTitle>`)
	contentBuilder.WriteString(`<text>` + ei.Title + "</text>")
	contentBuilder.WriteString(`</docTitle>`)
	contentBuilder.WriteString(`<navMap>`)

	if ei.coverImage != nil {
		playOrder++
		contentBuilder.WriteString(`<navPoint id="cover_page" playOrder="` + strconv.Itoa(playOrder) + `">`)
		contentBuilder.WriteString(`<navLabel>`)
		contentBuilder.WriteString(`<text>Cover</text>`)
		contentBuilder.WriteString(`</navLabel>`)
		contentBuilder.WriteString(`<content src="cover.xhtml" />`)
		contentBuilder.WriteString(`</navPoint>`)
	}

	playOrder++
	contentBuilder.WriteString(`<navPoint id="title_page" playOrder="` + strconv.Itoa(playOrder) + `">`)
	contentBuilder.WriteString(`<navLabel>`)
	contentBuilder.WriteString(`<text>Title</text>`)
	contentBuilder.WriteString(`</navLabel>`)
	contentBuilder.WriteString(`<content src="title.xhtml" />`)
	contentBuilder.WriteString(`</navPoint>`)

	playOrder++
	contentBuilder.WriteString(`<navPoint id="text_page" playOrder="` + strconv.Itoa(playOrder) + `">`)
	contentBuilder.WriteString(`<navLabel>`)
	contentBuilder.WriteString(`<text>Text</text>`)
	contentBuilder.WriteString(`</navLabel>`)
	contentBuilder.WriteString(`<content src="text.xhtml" />`)
	contentBuilder.WriteString(`</navPoint>`)

	contentBuilder.WriteString(`</navMap>`)
	contentBuilder.WriteString(`</ncx>`)

	if _, err = io.WriteString(w, contentBuilder.String()); err != nil {
		return
	}

	return
}

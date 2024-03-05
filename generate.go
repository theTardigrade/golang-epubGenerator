package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
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
		generateText,
		generateOCF,
		generateNCX,
	}
)

func generate(ei *epubInfo) (err error) {
	archiveFile, err := os.Create("epub.zip")
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
}

func generateMimetype(ei *epubInfo, archiveWriter *zip.Writer) (err error) {
	w, err := archiveWriter.Create("mimetype")
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
	b, err := os.ReadFile(ei.StylesPath)
	if err != nil {
		return
	}

	b, err = minifier.Bytes("text/css", b)
	if err != nil {
		return
	}

	w, err := archiveWriter.Create("styles.css")
	if err != nil {
		return
	}

	if _, err = w.Write(b); err != nil {
		return
	}

	return
}

func generateText(ei *epubInfo, archiveWriter *zip.Writer) (err error) {
	b, err := os.ReadFile(ei.ContentPath)
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

	if _, err = io.WriteString(w, xhtmlHeader(ei)); err != nil {
		return
	}

	if _, err = w.Write(b); err != nil {
		return
	}

	if _, err = io.WriteString(w, xhtmlFooter()); err != nil {
		return
	}

	fmt.Println(string(b))

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
	builder.WriteString(`<dc:identifier id="unique-id">` + ei.ISBN + `</dc:identifier>`)
	builder.WriteString(`</metadata>`)
	builder.WriteString(`<manifest>`)
	builder.WriteString(`<item id="styles" href="styles.css" media-type="text/css" />`)
	builder.WriteString(`<item id="text" href="text.xhtml" media-type="application/xhtml+xml" />`)
	builder.WriteString(`<item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml" />`)
	builder.WriteString(`</manifest>`)
	builder.WriteString(`<spine toc="ncx">`)
	builder.WriteString(`<itemref idref="text" />`)
	builder.WriteString(`</spine>`)
	// builder.WriteString(`<guide>`)
	// builder.WriteString(`</guide>`)
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
	contentBuilder.WriteString(`<navPoint id="text" playOrder="1">`)
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

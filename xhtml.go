package main

import "strings"

func xhtmlHeader(ei *epubInfo) string {
	var builder strings.Builder

	builder.WriteString(`<?xml version="1.0" encoding="utf-8"?>`)
	builder.WriteString(`<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en">`)
	builder.WriteString(`<head>`)
	builder.WriteString(`<title>` + ei.Title + `</title>`)
	builder.WriteString(`<link rel="stylesheet" href="styles.css" />`)
	builder.WriteString(`</head>`)
	builder.WriteString(`<body>`)

	return builder.String()
}

func xhtmlFooter() string {
	return "</body></html>"
}

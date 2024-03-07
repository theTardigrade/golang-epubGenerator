package main

import "strings"

func xhtmlHeader(title, headContent string) string {
	var builder strings.Builder

	builder.WriteString(`<?xml version="1.0" encoding="utf-8"?>`)
	builder.WriteString(`<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en">`)
	builder.WriteString(`<head>`)
	builder.WriteString(`<title>` + title + `</title>`)
	builder.WriteString(`<link rel="stylesheet" href="styles.css" />`)
	builder.WriteString(`<style>h1{page-break-before: always;}</style>`)

	if headContent != "" {
		builder.WriteString(headContent)
	}

	builder.WriteString(`</head>`)
	builder.WriteString(`<body>`)

	return builder.String()
}

func xhtmlFooter() string {
	return "</body></html>"
}

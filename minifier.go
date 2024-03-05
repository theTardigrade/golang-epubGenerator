package main

import (
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/xml"
)

var (
	minifier = minify.New()
)

func init() {
	minifier = minify.New()

	minifier.AddFunc("text/css", css.Minify)
	minifier.AddFunc("text/xml", xml.Minify)
}

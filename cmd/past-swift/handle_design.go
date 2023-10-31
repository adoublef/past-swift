package main

import (
	"embed"
	"net/http"

	"github.com/adoublef/past-swift/template"
)

//go:embed all:*.html
var embedFS embed.FS
var T = template.NewFS(embedFS, "*.html")

func handleDesign(t template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.ExecuteTemplate(w, "design.html", nil)
	}
}

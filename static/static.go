package static

import (
	"embed"
	"html/template"
	"net/http"
	"path"

	"github.com/benbjohnson/hashfs"
)

//go:embed all:css/*.css
var embedFS embed.FS
var hashFS = hashfs.NewFS(embedFS)

func Handler(prefix string) http.Handler {
	return http.StripPrefix(prefix, hashfs.FileServer(hashFS))
}

var FuncMap = template.FuncMap{
	"static": func(name string) string {
		return path.Join("/static", hashFS.HashName(name))
	},
}

/* 
func (s Static) ServeHTTP(w, r)
*/
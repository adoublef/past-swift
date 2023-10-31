package static

import (
	"embed"
	"html/template"
	"net/http"
	"path"

	"github.com/benbjohnson/hashfs"
)

var (
	//go:embed all:css/*.css all:js/*.js all:fonts/**/*.woff2
	embedFS embed.FS
	hashFS  = hashfs.NewFS(embedFS)
)

func Handler(prefix string) http.Handler {
	return http.StripPrefix(prefix, hashfs.FileServer(hashFS))
}

var FuncMap = func(prefix string) template.FuncMap {
	// doOnce maybe
	return template.FuncMap{
		"static": func(name string) string {
			return path.Join(prefix, hashFS.HashName(name))
		},
	}
}

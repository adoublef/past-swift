package template

/*
	https://charly3pins.dev/blog/learn-how-to-use-the-embed-package-in-go-by-building-a-web-page-easily/
	https://www.josephspurrier.com/how-to-embed-assets-in-go-1-16
	https://osinet.fr/go/en/articles/bundling-templates-with-embed/
	https://lets-go.alexedwards.net/sample/02.07-html-templating-and-inheritance.html
	https://blog.jetbrains.com/go/2022/11/08/build-a-blog-with-go-templates/#creating-a-new-project
	https://pliutau.com/struct-render-helper/
*/
import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type FS struct {
	fsys     fs.FS
	root     string
	t        *template.Template
	patterns []string
}

func NewFS(fsys fs.FS, root string, patterns ...string) *FS {
	// patterns go here
	return &FS{fsys: fsys, root: root, t: template.New("")}
}

func (fsys *FS) Funcs(funcMap template.FuncMap) *FS {
	fsys.t.Funcs(template.FuncMap(funcMap))
	return fsys
}

func (fsys *FS) Parse() (*template.Template, error) {
	patterns := []string{filepath.Join(fsys.root, "*.html")}
	// make `builtins` doOnce
	return fsys.t.Funcs(builtins()).ParseFS(fsys.fsys, patterns...)
}

type Template struct {
	t *template.Template
}

func (t *Template) Execute(wr io.Writer, name string, data any) error {
	return t.t.ExecuteTemplate(wr, name, data)
}

// func (fsys *FS) ExecuteTemplate() {}

// builtins adds default functions to be used in a template.
//
//   - Map allows for a map to be passed into the pipeline inline of a template.
var builtins = func() template.FuncMap {
	return template.FuncMap{
		"map": func(pairs ...any) (map[string]any, error) {
			if len(pairs)%2 != 0 {
				return nil, errors.New("misaligned map")
			}
			m := make(map[string]any, len(pairs)/2)

			for i := 0; i < len(pairs); i += 2 {
				key, ok := pairs[i].(string)
				if !ok {
					return nil, fmt.Errorf("cannot use type %T as map key", pairs[i])
				}
				m[key] = pairs[i+1]
			}
			return m, nil
		},
		"env": func(key string) string {
			return os.Getenv(key)
		},
	}
}

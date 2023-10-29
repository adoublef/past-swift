package fly

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func Replay(dsn string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if instance, err := lookUpPrimary(dsn); err != nil {
				h.ServeHTTP(w, r)
			} else {
				log.Printf("redirecting to primary instance: %q\n", instance)
				w.Header().Set("fly-replay", fmt.Sprintf("instance=%s", instance))
			}
		})
	}
}

func lookUpPrimary(dsn string) (string, error) {
	filename := filepath.Join(filepath.Dir(dsn), ".primary")
	primary, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(primary), nil
}

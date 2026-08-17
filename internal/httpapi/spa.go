package httpapi

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/robin-vidal/neetrack/internal/httpapi/assets"
)

// spaHandler serves the embedded frontend, falling back to index.html for
// any path that isn't a real static file so client-side routes resolve.
//
// fs.Sub can only fail if the dist/ subtree is missing from the embedded
// FS, which the //go:embed all:dist directive guarantees at compile time.
func spaHandler() http.Handler {
	dist, err := fs.Sub(assets.FS, "dist")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServerFS(dist)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if _, err := fs.Stat(dist, path); err != nil {
			// "/" resolves to index.html without a redirect; the explicit
			// "/index.html" path does not, since http.FileServer treats
			// it as the canonical index and redirects away from it.
			r = cloneRequestPath(r, "/")
		}

		fileServer.ServeHTTP(w, r)
	})
}

func cloneRequestPath(r *http.Request, path string) *http.Request {
	clone := r.Clone(r.Context())
	clone.URL.Path = path
	return clone
}

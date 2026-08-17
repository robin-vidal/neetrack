// Package assets embeds the built frontend. The Docker build copies web/dist
// into dist/ here before compiling; outside that build, dist/ holds only a
// placeholder so the package still compiles.
package assets

import "embed"

//go:embed all:dist
var FS embed.FS

// Package migrations embeds the goose SQL migration files into the binary.
package migrations

import "embed"

// FS holds every *.sql migration file, applied at boot by internal/store.
//
//go:embed *.sql
var FS embed.FS

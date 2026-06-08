// Package db embeds the SQL migration files so the distroless production binary
// ships them (go:embed cannot reach outside the module, hence migrations live
// here under apps/api rather than at the repo root).
package db

import "embed"

// FS holds the goose migration files.
//
//go:embed migrations/*.sql
var FS embed.FS

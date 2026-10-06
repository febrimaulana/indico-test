// Package migrations embeds the SQL migration files so the backend binary is
// fully self-contained and can migrate on startup with no external tooling.
package migrations

import "embed"

// FS contains every .up.sql / .down.sql migration file.
//
//go:embed *.sql
var FS embed.FS

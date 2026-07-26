package migrations

import "embed"

// FS — SQL-миграции, вкомпилированные в бинарь.
//
//go:embed *.sql
var FS embed.FS

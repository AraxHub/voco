// Package swaggerdocs holds generated OpenAPI docs (swag).
// Regenerating: see local docs/swagger.md (folder docs/ is gitignored).
package swaggerdocs

import "github.com/swaggo/swag"

const docTemplate = `{
    "swagger": "2.0",
    "info": {
        "title": "Voco API",
        "description": "Rooms, chat, calendar, push",
        "version": "1.0"
    },
    "basePath": "/api/v1",
    "paths": {}
}`

type s struct{}

func (s *s) ReadDoc() string { return docTemplate }

func init() {
	swag.Register(swag.Name, &s{})
}

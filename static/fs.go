package ui

import "embed"

//go:embed *
var FS embed.FS

const EmbedFileName = "fs.go"

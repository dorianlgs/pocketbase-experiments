package ui

import (
	"embed"
	"io/fs"
)

//go:generate npm install
//go:generate npm run build
//go:embed all:build
var distDir embed.FS

var DistDirFS, _ = fs.Sub(distDir, "build")

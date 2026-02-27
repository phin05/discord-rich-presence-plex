package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFs embed.FS

var BuildOutput = func() fs.FS {
	buildOutput, err := fs.Sub(distFs, "dist")
	if err != nil {
		panic(err)
	}
	return buildOutput
}()

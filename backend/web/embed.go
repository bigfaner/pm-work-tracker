package web

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed all:dist
var FS embed.FS

// ValidateAssets checks that the embedded FS contains dist/index.html.
func ValidateAssets(fsys embed.FS) error {
	return ValidateAssetsFS(fsys)
}

// ValidateAssetsFS accepts any fs.FS, making it testable with fstest.MapFS.
func ValidateAssetsFS(fsys fs.FS) error {
	_, err := fsys.Open("dist/index.html")
	if err != nil {
		return fmt.Errorf("embedded assets missing dist/index.html — did the build script succeed?")
	}
	return nil
}

package web

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAssets_Missing(t *testing.T) {
	fsys := fstest.MapFS{}
	err := ValidateAssetsFS(fsys)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dist/index.html")
}

func TestValidateAssets_Present(t *testing.T) {
	fsys := fstest.MapFS{
		"dist/index.html": &fstest.MapFile{Data: []byte("<html></html>")},
	}
	err := ValidateAssetsFS(fsys)
	assert.NoError(t, err)
}

package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeStatic_Found(t *testing.T) {
	fsys := fstest.MapFS{
		"dist/assets/app.js": &fstest.MapFile{Data: []byte("console.log('hello')")},
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "filepath", Value: "/app.js"}}

	ServeStatic(fsys)(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "max-age=31536000, immutable", w.Header().Get("Cache-Control"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "console.log('hello')", w.Body.String())
}

func TestServeStatic_NotFound(t *testing.T) {
	fsys := fstest.MapFS{}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "filepath", Value: "/missing.js"}}

	ServeStatic(fsys)(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), `"not found"`)
}

func TestServeSPA(t *testing.T) {
	fsys := fstest.MapFS{
		"dist/index.html": &fstest.MapFile{Data: []byte("<html><body>app</body></html>")},
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ServeSPA(fsys)(c)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
	assert.Equal(t, "<html><body>app</body></html>", w.Body.String())
}

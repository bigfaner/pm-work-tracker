package handler

import (
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// ServeStatic serves files from dist/assets/ with long-term cache headers.
func ServeStatic(fsys fs.FS) gin.HandlerFunc {
	return func(c *gin.Context) {
		filePath := "dist/assets/" + strings.TrimPrefix(c.Param("filepath"), "/")
		f, err := fsys.Open(filePath)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		defer f.Close()

		c.Header("Cache-Control", "max-age=31536000, immutable")
		c.Header("X-Content-Type-Options", "nosniff")

		contentType := mime.TypeByExtension(filepath.Ext(filePath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		data, err := io.ReadAll(f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "read error"})
			return
		}

		c.Data(http.StatusOK, contentType, data)
	}
}

// ServeSPA always returns dist/index.html with no-cache headers.
func ServeSPA(fsys fs.FS) gin.HandlerFunc {
	return func(c *gin.Context) {
		f, err := fsys.Open("dist/index.html")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		defer f.Close()

		data, err := io.ReadAll(f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "read error"})
			return
		}

		c.Header("Cache-Control", "no-cache")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	}
}

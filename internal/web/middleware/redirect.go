package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RedirectMiddleware(basePath string) gin.HandlerFunc {
	basePath = normalizeBasePath(basePath)

	type redirectRule struct {
		from string
		to   string
	}

	// Keep routes ordered from most-specific to least-specific.
	redirects := []redirectRule{
		{from: "panel/API", to: "panel/api"},
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		for _, rule := range redirects {
			from, to := basePath+rule.from, basePath+rule.to

			if hasSegmentPrefix(path, from) {
				newPath := to + path[len(from):]
				if rawQuery := c.Request.URL.RawQuery; rawQuery != "" {
					newPath += "?" + rawQuery
				}

				c.Redirect(http.StatusPermanentRedirect, newPath)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func hasSegmentPrefix(path string, prefix string) bool {
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	if len(path) == len(prefix) {
		return true
	}
	return path[len(prefix)] == '/'
}

func normalizeBasePath(basePath string) string {
	if basePath == "" {
		return "/"
	}
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}
	return basePath
}

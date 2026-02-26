package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHasSegmentPrefix(t *testing.T) {
	t.Run("exact match", func(t *testing.T) {
		if !hasSegmentPrefix("/panel/API", "/panel/API") {
			t.Fatal("expected exact match to be true")
		}
	})

	t.Run("child segment", func(t *testing.T) {
		if !hasSegmentPrefix("/panel/API/status", "/panel/API") {
			t.Fatal("expected child segment to be true")
		}
	})

	t.Run("similar prefix must not match", func(t *testing.T) {
		if hasSegmentPrefix("/panel/APIs", "/panel/API") {
			t.Fatal("expected similar prefix to be false")
		}
	})
}

func TestNormalizeBasePath(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: "", want: "/"},
		{in: "/", want: "/"},
		{in: "ahmand", want: "/ahmand/"},
		{in: "/ahmand", want: "/ahmand/"},
		{in: "/ahmand/", want: "/ahmand/"},
	}

	for _, tc := range cases {
		if got := normalizeBasePath(tc.in); got != tc.want {
			t.Fatalf("normalizeBasePath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestRedirectMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(RedirectMiddleware("/"))
	engine.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	t.Run("panel API case rewrite", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/panel/API/status?a=1&b=2", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusPermanentRedirect {
			t.Fatalf("expected 308, got %d", w.Code)
		}
		if got := w.Header().Get("Location"); got != "/panel/api/status?a=1&b=2" {
			t.Fatalf("unexpected redirect location: %s", got)
		}
	})

	t.Run("panel API POST uses method-preserving redirect", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/panel/API/status", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusPermanentRedirect {
			t.Fatalf("expected 308, got %d", w.Code)
		}
		if got := w.Header().Get("Location"); got != "/panel/api/status" {
			t.Fatalf("unexpected redirect location: %s", got)
		}
	})

	t.Run("legacy xui must not redirect", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/xui/API/status", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 (no route), got %d", w.Code)
		}
	})

	t.Run("legacy panel v2 must not redirect", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/panel/v2/inbounds", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 (no route), got %d", w.Code)
		}
	})
}

func TestRedirectMiddlewareWithCustomBasePath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(RedirectMiddleware("/ahmand/"))
	engine.GET("/ahmand/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	t.Run("custom base panel API", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ahmand/panel/API/status?x=1", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusPermanentRedirect {
			t.Fatalf("expected 308, got %d", w.Code)
		}
		if got := w.Header().Get("Location"); got != "/ahmand/panel/api/status?x=1" {
			t.Fatalf("unexpected redirect location: %s", got)
		}
	})

	t.Run("custom base should not rewrite xui path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ahmand/xui/API/status", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 (no route), got %d", w.Code)
		}
	})
}

func TestRedirectMiddlewareWithNonCanonicalBasePathInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(RedirectMiddleware("ahmand"))

	req := httptest.NewRequest(http.MethodGet, "/ahmand/panel/API/status?x=1", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusPermanentRedirect {
		t.Fatalf("expected 308, got %d", w.Code)
	}
	if got := w.Header().Get("Location"); got != "/ahmand/panel/api/status?x=1" {
		t.Fatalf("unexpected redirect location: %s", got)
	}
}

package webapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/labstack/echo/v4"
)

func TestSPAFallbackHandler(t *testing.T) {
	// モックファイルシステム
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><body>SPA Root</body></html>"),
		},
		"map/index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><body>Map Page</body></html>"),
		},
		"_nuxt/app.js": &fstest.MapFile{
			Data: []byte("console.log('app');"),
		},
	}

	e := echo.New()
	handler := spaHandler(http.FS(mockFS))

	// 1. ルートへのアクセス -> index.html
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error for /: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "SPA Root") {
		t.Errorf("expected body to contain 'SPA Root', got %s", rec.Body.String())
	}

	// 2. 実在する静的ファイルへのアクセス -> _nuxt/app.js
	req = httptest.NewRequest(http.MethodGet, "/_nuxt/app.js", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error for /_nuxt/app.js: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "console.log('app');") {
		t.Errorf("expected body to contain app.js code, got %s", rec.Body.String())
	}

	// 3. 実在するサブディレクトリへのアクセス -> /map
	req = httptest.NewRequest(http.MethodGet, "/map", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error for /map: %v", err)
	}
	// /map は 200 または 301 (ディレクトリ末尾スラッシュリダイレクト)
	if rec.Code != http.StatusOK && rec.Code != http.StatusMovedPermanently {
		t.Errorf("expected 200 or 301, got %d", rec.Code)
	}

	// 4. 動的ルート（静的ファイルが存在しない）へのアクセス -> /dashboard/488115978f44609bd7eaaf63a085b807a67e9d300cbc1f8c
	req = httptest.NewRequest(http.MethodGet, "/dashboard/488115978f44609bd7eaaf63a085b807a67e9d300cbc1f8c", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error for dynamic route: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK for SPA fallback, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "SPA Root") {
		t.Errorf("expected body to contain 'SPA Root' from index.html fallback, got %s", rec.Body.String())
	}

	// 5. 存在しない静的アセット（.png）へのアクセス -> 404
	req = httptest.NewRequest(http.MethodGet, "/images/not-found.png", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error for missing image: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for missing asset, got %d", rec.Code)
	}
}

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Brunlx/BLX/internal/static"
	"github.com/Brunlx/BLX/internal/tools"
)

func newTestServer() *Server {
	return New(tools.NewCatalog())
}

func doRaw(t *testing.T, h http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("erro ao codificar body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	h := newTestServer().Routes(nil)
	rec := doJSON(t, h, http.MethodGet, "/api/health", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obtive %d", rec.Code)
	}
}

func TestListTools(t *testing.T) {
	h := newTestServer().Routes(nil)
	rec := doJSON(t, h, http.MethodGet, "/api/tools", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obtive %d", rec.Code)
	}
	var payload struct {
		Categories []categoryInfo `json:"categories"`
		Tools      []*tools.Tool  `json:"tools"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json inválido: %v", err)
	}
	if len(payload.Tools) < 10 {
		t.Errorf("esperava >= 10 ferramentas, obtive %d", len(payload.Tools))
	}
	if len(payload.Categories) < 5 {
		t.Errorf("esperava >= 5 categorias, obtive %d", len(payload.Categories))
	}
}

func TestGetToolNotFound(t *testing.T) {
	h := newTestServer().Routes(nil)
	rec := doJSON(t, h, http.MethodGet, "/api/tools/nao-existe", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404, obtive %d", rec.Code)
	}
}

func TestGenerateSuccess(t *testing.T) {
	h := newTestServer().Routes(nil)
	rec := doJSON(t, h, http.MethodPost, "/api/generate", map[string]any{
		"toolId": "nmap",
		"answers": map[string]string{
			"targets":  "192.168.1.0/24",
			"scanType": "aggr",
			"verbose":  "true",
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obtive %d (body: %s)", rec.Code, rec.Body.String())
	}
	var res tools.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("json inválido: %v", err)
	}
	if res.ToolID != "nmap" || len(res.Commands) == 0 {
		t.Errorf("resultado incompleto: %+v", res)
	}
	if !bytes.Contains([]byte(res.Commands[0].Code), []byte("-A")) {
		t.Errorf("esperava -A no comando, obtive %q", res.Commands[0].Code)
	}
}

func TestGenerateValidationError(t *testing.T) {
	h := newTestServer().Routes(nil)
	rec := doJSON(t, h, http.MethodPost, "/api/generate", map[string]any{
		"toolId":  "nmap",
		"answers": map[string]string{},
	})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("esperava 422, obtive %d", rec.Code)
	}
	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json inválido: %v", err)
	}
	if payload["id"] == "" {
		t.Error("422 deve incluir o id da pergunta que falhou")
	}
	if payload["question"] == "" || payload["reason"] == "" {
		t.Error("422 deve incluir question e reason")
	}
}

func TestUnknownAPIEndpointReturnsJSON404(t *testing.T) {
	h := newTestServer().Routes(nil)
	rec := doJSON(t, h, http.MethodGet, "/api/nao-existe", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404, obtive %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("404 de /api/ deveria ser JSON, Content-Type=%q", ct)
	}
	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("404 de /api/ deveria ter corpo JSON válido: %v", err)
	}
}

func TestAPIResponsesAreNotCached(t *testing.T) {
	h := newTestServer().Routes(nil)
	for _, path := range []string{"/api/health", "/api/tools"} {
		rec := doJSON(t, h, http.MethodGet, path, nil)
		if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
			t.Errorf("%s: esperava Cache-Control: no-store, obtive %q", path, cc)
		}
	}
	rec := doJSON(t, h, http.MethodPost, "/api/generate", map[string]any{
		"toolId":  "nmap",
		"answers": map[string]string{},
	})
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("generate 422: esperava Cache-Control: no-store, obtive %q", cc)
	}
}

func TestGenerateMalformedBody(t *testing.T) {
	h := newTestServer().Routes(nil)
	rec := doRaw(t, h, http.MethodPost, "/api/generate", []byte("{isso não é json"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obtive %d", rec.Code)
	}
}

func TestUnknownToolGenerate(t *testing.T) {
	h := newTestServer().Routes(nil)
	rec := doJSON(t, h, http.MethodPost, "/api/generate", map[string]any{
		"toolId":  "nada",
		"answers": map[string]string{},
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404, obtive %d", rec.Code)
	}
}

func TestStaticServed(t *testing.T) {
	staticH, err := static.Handler()
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	h := newTestServer().Routes(staticH)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200 na raiz, obtive %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("<!doctype html>")) && !bytes.Contains(rec.Body.Bytes(), []byte("<!DOCTYPE")) {
		t.Errorf("raiz deveria servir o index.html, obtive: %.100s", rec.Body.String())
	}
}

func TestAPISlashExactReturnsJSON404(t *testing.T) {
	h := newTestServer().Routes(nil)
	rec := doJSON(t, h, http.MethodGet, "/api", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404, obtive %d", rec.Code)
	}
	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("/api deveria ter corpo JSON válido: %v", err)
	}
}

func TestPanicRecoveryReturns500(t *testing.T) {
	s := newTestServer()
	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	h := s.withMiddleware(panicking)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/boom", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("esperava 500 após panic, obtive %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("erro interno")) {
		t.Errorf("esperava mensagem de erro JSON, obtive: %s", rec.Body.String())
	}
}

func TestCORSOnlyRepliesToRealPreflight(t *testing.T) {
	s := New(tools.NewCatalog())
	s.corsOrigin = "http://example.com"
	h := s.Routes(nil)

	// OPTIONS com Access-Control-Request-Method em caminho /api/ → 204
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/tools", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("esperava 204 para preflight real, obtive %d", rec.Code)
	}

	// OPTIONS sem pedido de CORS em caminho /api/ → 404 (não deve responder 204)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodOptions, "/api/tools", nil)
	req.Header.Set("Origin", "http://example.com")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("OPTIONS sem Access-Control-Request-Method não deveria receber 204, obtive %d", rec.Code)
	}
}

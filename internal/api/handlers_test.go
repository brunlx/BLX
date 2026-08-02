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

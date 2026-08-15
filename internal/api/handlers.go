package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Brunlx/BLX/internal/tools"
)

// categoryInfo carries a category id and its display name.
type categoryInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var categoryNames = map[string]string{
	tools.CatRecon:       "Reconhecimento",
	tools.CatWeb:         "Web",
	tools.CatExploit:     "Exploração",
	tools.CatCreds:       "Credenciais",
	tools.CatWireless:    "Wireless",
	tools.CatNetwork:     "Tráfego de Rede",
	tools.CatPostExploit: "Pós-exploração / AD",
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleListTools(w http.ResponseWriter, r *http.Request) {
	tools := s.catalog.Tools()
	categories := make([]categoryInfo, 0, len(categoryNames))
	seen := map[string]bool{}
	for _, t := range tools {
		if !seen[t.Category] {
			seen[t.Category] = true
			categories = append(categories, categoryInfo{ID: t.Category, Name: categoryNames[t.Category]})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"categories": categories,
		"tools":      tools,
	})
}

func (s *Server) handleGetTool(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tool := s.catalog.Tool(id)
	if tool == nil {
		writeError(w, http.StatusNotFound, "ferramenta não encontrada")
		return
	}
	writeJSON(w, http.StatusOK, tool)
}

// generateRequest is the payload sent by the frontend wizard.
type generateRequest struct {
	ToolID  string            `json:"toolId"`
	Answers map[string]string `json:"answers"`
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	var req generateRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo de requisição inválido")
		return
	}
	if req.ToolID == "" {
		writeError(w, http.StatusBadRequest, "informe o campo toolId")
		return
	}
	if req.Answers == nil {
		req.Answers = map[string]string{}
	}

	result, err := s.catalog.Generate(req.ToolID, req.Answers)
	if err != nil {
		var ve *tools.ValidationError
		if errors.As(err, &ve) {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{
				"error":    "validação falhou",
				"id":       ve.ID,
				"question": ve.Question,
				"reason":   ve.Reason,
			})
			return
		}
		if errors.Is(err, tools.ErrUnknownTool) {
			writeError(w, http.StatusNotFound, "ferramenta não encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao gerar comandos")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

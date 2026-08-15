// Package tools defines the catalog of pentesting tools, their configuration
// question flows and the engines that turn answers into ready-to-run commands.
package tools

import (
	"strconv"
	"strings"
)

// Option is a single selectable value inside a select/multi question.
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// Question represents one configuration prompt shown to the operator.
// Supported types: text, number, select, multi, boolean.
type Question struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Placeholder string   `json:"placeholder,omitempty"`
	Help        string   `json:"help,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Options     []Option `json:"options,omitempty"`
	Default     string   `json:"default,omitempty"`
	Min         int      `json:"-"`
	Max         int      `json:"-"`
}

// Tool describes a pentesting tool available in the catalog.
type Tool struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Icon        string     `json:"icon"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Risk        string     `json:"risk"` // low | medium | high
	Install     string     `json:"install"`
	Docs        string     `json:"docs"`
	Tags        []string   `json:"tags"`
	Questions   []Question `json:"questions"`
}

// Command is a single generated snippet ready to be copied by the operator.
type Command struct {
	Title    string `json:"title"`
	Code     string `json:"code"`
	Language string `json:"language"`
	Hint     string `json:"hint,omitempty"`
}

// Result is the full output of a generation run.
type Result struct {
	ToolID   string    `json:"toolId"`
	ToolName string    `json:"toolName"`
	Commands []Command `json:"commands"`
	Notes    []string  `json:"notes"`
	Warnings []string  `json:"warnings"`
	Risk     string    `json:"risk"`
}

// Generator turns a validated set of answers into ready-to-run commands.
type Generator func(tool *Tool, answers map[string]string) (*Result, error)

// Catalog is an immutable registry of tools and their generators.
type Catalog struct {
	tools      []*Tool
	generators map[string]Generator
}

// NewCatalog builds a catalog from the built-in definitions.
func NewCatalog() *Catalog {
	c := &Catalog{generators: make(map[string]Generator)}
	c.registerAll()
	return c
}

// Tools returns all tools in registration order.
func (c *Catalog) Tools() []*Tool { return c.tools }

// Tool returns a single tool by ID, or nil.
func (c *Catalog) Tool(id string) *Tool {
	for _, t := range c.tools {
		if t.ID == id {
			return t
		}
	}
	return nil
}

// Validate checks that required answers are present and that select/multi
// answers reference valid options.
func (c *Catalog) Validate(id string, answers map[string]string) error {
	t := c.Tool(id)
	if t == nil {
		return ErrUnknownTool
	}
	for _, q := range t.Questions {
		val, ok := answers[q.ID]
		empty := !ok || strings.TrimSpace(val) == ""
		if empty && q.Required {
			return &ValidationError{ID: q.ID, Question: q.Label, Reason: "campo obrigatório não preenchido"}
		}
		if empty {
			continue
		}
		if q.Type == "number" {
			n, err := strconv.Atoi(strings.TrimSpace(val))
			if err != nil {
				return &ValidationError{ID: q.ID, Question: q.Label, Reason: "informe um número inteiro válido"}
			}
			if (q.Min != 0 && n < q.Min) || (q.Max != 0 && n > q.Max) {
				return &ValidationError{ID: q.ID, Question: q.Label, Reason: "valor fora do intervalo permitido"}
			}
		}
		if q.Type == "select" && !hasOption(q, val) {
			return &ValidationError{ID: q.ID, Question: q.Label, Reason: "valor inválido"}
		}
		if q.Type == "multi" && !hasAllOptions(q, val) {
			return &ValidationError{ID: q.ID, Question: q.Label, Reason: "valor inválido"}
		}
	}
	return nil
}

// Generate validates the answers and delegates to the tool generator.
// Unanswered optional questions that declare a default are filled in first,
// so generators never see an empty value where a sensible default exists
// (e.g. gobuster "dir", impacket "psexec", mimikatz "logonpasswords").
func (c *Catalog) Generate(id string, answers map[string]string) (*Result, error) {
	t := c.Tool(id)
	if t == nil {
		return nil, ErrUnknownTool
	}
	norm := withDefaults(t, answers)
	if err := c.Validate(id, norm); err != nil {
		return nil, err
	}
	gen, ok := c.generators[id]
	if !ok {
		return nil, ErrNoGenerator
	}
	return gen(t, norm)
}

// withDefaults returns a copy of answers where unanswered questions that
// declare a default value are filled in. The caller's map is left untouched.
func withDefaults(t *Tool, answers map[string]string) map[string]string {
	norm := make(map[string]string, len(answers)+len(t.Questions))
	for k, v := range answers {
		norm[k] = v
	}
	for _, q := range t.Questions {
		if q.Default == "" {
			continue
		}
		if v, ok := norm[q.ID]; !ok || strings.TrimSpace(v) == "" {
			norm[q.ID] = q.Default
		}
	}
	return norm
}

func hasOption(q Question, val string) bool {
	for _, o := range q.Options {
		if o.Value == val {
			return true
		}
	}
	return false
}

func hasAllOptions(q Question, val string) bool {
	for _, v := range splitCSV(val) {
		if !hasOption(q, v) {
			return false
		}
	}
	return true
}

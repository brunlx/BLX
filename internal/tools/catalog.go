// Package tools defines the catalog of pentesting tools, their configuration
// question flows and the engines that turn answers into ready-to-run commands.
package tools

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
		empty := !ok || val == ""
		if empty && q.Required {
			return &ValidationError{Question: q.Label, Reason: "campo obrigatório não preenchido"}
		}
		if empty {
			continue
		}
		if q.Type == "select" && !hasOption(q, val) {
			return &ValidationError{Question: q.Label, Reason: "valor inválido: " + val}
		}
		if q.Type == "multi" && !hasAllOptions(q, val) {
			return &ValidationError{Question: q.Label, Reason: "valor inválido: " + val}
		}
	}
	return nil
}

// Generate validates the answers and delegates to the tool generator.
func (c *Catalog) Generate(id string, answers map[string]string) (*Result, error) {
	t := c.Tool(id)
	if t == nil {
		return nil, ErrUnknownTool
	}
	if err := c.Validate(id, answers); err != nil {
		return nil, err
	}
	gen, ok := c.generators[id]
	if !ok {
		return nil, ErrNoGenerator
	}
	return gen(t, answers)
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

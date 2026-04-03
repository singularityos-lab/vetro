package lsp

import (
	"encoding/json"
	"sort"
	"strings"
	"vetro/internal/domain/vetro"
)

type DidOpenOrChangeParams struct {
	TextDocument struct {
		URI  string `json:"uri"`
		Text string `json:"text"`
	} `json:"textDocument"`
	ContentChanges []struct {
		Text string `json:"text"`
	} `json:"contentChanges"`
}

type CompletionParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position Position `json:"position"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Message  string `json:"message"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

func (s *Server) handleDidOpenOrChange(paramsRaw json.RawMessage) {
	var params DidOpenOrChangeParams
	json.Unmarshal(paramsRaw, &params)

	uri := params.TextDocument.URI
	var text string
	if len(params.ContentChanges) > 0 {
		text = params.ContentChanges[0].Text
	} else {
		text = params.TextDocument.Text
	}
	s.files[uri] = text

	// Run diagnostics
	s.publishDiagnostics(uri, text)
}

func (s *Server) publishDiagnostics(uri string, text string) {
	lexer := vetro.NewLexer(text)
	tokens, _ := lexer.Tokenize()
	parser := vetro.NewParser(tokens)
	_, err := parser.Parse()

	diagnostics := []Diagnostic{}
	if err != nil {
		if pErr, ok := err.(*vetro.ParserError); ok {
			line := pErr.Token.Line
			if line > 0 {
				line-- // LSP uses 0-based lines
			}
			diagnostics = append(diagnostics, Diagnostic{
				Range: Range{
					Start: Position{Line: line, Character: 0},
					End:   Position{Line: line, Character: 100},
				},
				Severity: 1, // Error
				Message:  pErr.Message,
			})
		}
	}

	s.notify("textDocument/publishDiagnostics", map[string]any{
		"uri":         uri,
		"diagnostics": diagnostics,
	})
}

func (s *Server) handleCompletion(id any, paramsRaw json.RawMessage) {
	var params CompletionParams
	json.Unmarshal(paramsRaw, &params)

	text := s.files[params.TextDocument.URI]
	lines := strings.Split(text, "\n")
	if params.Position.Line >= len(lines) {
		s.respond(id, []any{})
		return
	}
	line := lines[params.Position.Line]
	char := params.Position.Character
	if char < 0 {
		char = 0
	}
	if char > len(line) {
		char = len(line)
	}
	prefix := line[:char]

	items := []map[string]any{}

	// Simple context detection
	if strings.Contains(prefix, ".") {
		// Modifier/Signal/Property completion
		if vetro.MetadataManager != nil && vetro.MetadataManager.Metadata != nil {
			seen := make(map[string]bool)
			for _, class := range vetro.MetadataManager.Metadata.Classes {
				for prop := range class.Properties {
					if !seen[prop] {
						items = append(items, map[string]any{
							"label": prop,
							"kind":  10, // Property
						})
						seen[prop] = true
					}
				}
				for _, sig := range class.Signals {
					if sig == "" {
						continue
					}
					label := "on" + strings.ToUpper(sig[:1]) + sig[1:]
					if !seen[label] {
						items = append(items, map[string]any{
							"label": label,
							"kind":  11, // Event/Signal
						})
						seen[label] = true
					}
				}
			}
		}
	} else {
		// Component completion
		for _, comp := range getCoreComponents() {
			items = append(items, map[string]any{
				"label": comp,
				"kind":  7, // Class
			})
		}
	}

	s.respond(id, items)
}

func getCoreComponents() []string {
	if vetro.MetadataManager != nil && vetro.MetadataManager.Metadata != nil {
		seen := make(map[string]bool)
		components := make([]string, 0, len(vetro.MetadataManager.Metadata.Classes))

		for className := range vetro.MetadataManager.Metadata.Classes {
			component := strings.TrimPrefix(className, "Gtk")
			if component == "" || component == className {
				continue
			}
			if !seen[component] {
				components = append(components, component)
				seen[component] = true
			}
		}

		if len(components) > 0 {
			sort.Strings(components)
			return components
		}
	}

	// Fallback to a hardcoded list of common GTK components if metadata is unavailable,
	// not the most elegant solution but ensures some level of completion support even
	// without GIR metadata.
	return []string{
		"Window", "ApplicationWindow", "Box", "VBox", "HBox",
		"Label", "Button", "Entry", "Image", "Stack", "Grid",
		"HeaderBar", "MenuButton", "ScrolledWindow", "ListBox",
	}
}

package action

import (
	"fmt"
	"path/filepath"
	"strings"

	"vetro/internal/domain/vetro"
)

// TranspileAction orchestrates the transpilation process.
type TranspileAction struct{}

// Transpile reads an input file and writes to the output file.
// It auto-detects the conversion direction based on file extensions.
func (a *TranspileAction) Transpile(inPath, outPath string) error {
	inExt := strings.ToLower(filepath.Ext(inPath))
	outExt := strings.ToLower(filepath.Ext(outPath))

	if inExt == ".ui" && outExt == ".vetro" {
		return a.uiToVetro(inPath, outPath)
	}
	return a.vetroToUI(inPath, outPath)
}

// vetroToUI converts a .vetro file to a GTK .ui.
func (a *TranspileAction) vetroToUI(inPath, outPath string) error {
	source, err := vetro.ReadFile(inPath)
	if err != nil {
		return err
	}

	// Lex
	l := vetro.NewLexer(source)
	tokens, err := l.Tokenize()
	if err != nil {
		return err
	}

	// Parse
	p := vetro.NewParser(tokens)
	program, err := p.Parse()
	if err != nil {
		return err
	}

	// Generate XML
	g := vetro.NewGenerator()
	xmlOutput, err := g.Generate(program)
	if err != nil {
		return err
	}

	// Write the output file
	if err := vetro.WriteFile(outPath, xmlOutput); err != nil {
		return err
	}

	// Generate and write CSS file if there are style blocks
	cssOutput := g.GenerateCSS(program)
	if cssOutput != "" {
		cssPath := strings.TrimSuffix(outPath, filepath.Ext(outPath)) + ".css"
		if err := vetro.WriteFile(cssPath, cssOutput); err != nil {
			return fmt.Errorf("failed to write CSS file %s: %w", cssPath, err)
		}
	}

	return nil
}

// uiToVetro converts a GTK .ui file to Vetro source.
func (a *TranspileAction) uiToVetro(inPath, outPath string) error {
	xmlSource, err := vetro.ReadFile(inPath)
	if err != nil {
		return err
	}

	emitter := vetro.NewVetroEmitter()
	vetroSource, err := emitter.EmitXMLToVetro(xmlSource)
	if err != nil {
		return err
	}

	return vetro.WriteFile(outPath, vetroSource)
}

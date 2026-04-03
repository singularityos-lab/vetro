package vetro

import (
	"strings"
	"testing"
)

func TestGenerator_SimpleXML(t *testing.T) {
	input := `Window(id: "main", title: "Test") {}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, _ := parser.Parse()
	generator := NewGenerator()
	output, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<interface>`,
		`<object class="GtkWindow" id="main">`,
		`<property name="title">Test</property>`,
		`</interface>`,
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("output missing: %s", check)
		}
	}
}

func TestGenerator_BooleanOutput(t *testing.T) {
	input := `Window(resizable: false) {}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, _ := parser.Parse()
	generator := NewGenerator()
	output, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `<property name="resizable">false</property>`) {
		t.Error("boolean false not properly formatted in XML")
	}
}

func TestGenerator_Deterministic(t *testing.T) {
	input := `Window(id: "main") {
		VBox(spacing: 10) {
			Label("A")
			Button("B")
		}
	}`

	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, _ := parser.Parse()
	generator := NewGenerator()

	var outputs []string
	for i := 0; i < 5; i++ {
		output, err := generator.Generate(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		outputs = append(outputs, output)
	}

	for i := 1; i < len(outputs); i++ {
		if outputs[i] != outputs[0] {
			t.Errorf("output %d differs from first output", i)
		}
	}
}

func TestGenerator_XMLEscaping(t *testing.T) {
	input := `Label("Test & <value>") {}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, _ := parser.Parse()
	generator := NewGenerator()
	output, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Test &amp; &lt;value&gt;") {
		t.Error("XML escaping not working properly")
	}
}

func TestGenerator_SignalGeneration(t *testing.T) {
	input := `Button("Click").onClicked(handler) {}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, _ := parser.Parse()
	generator := NewGenerator()
	output, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `<signal name="clicked" handler="handler" swapped="no"/>`) {
		t.Error("signal not properly generated")
	}
}

func TestGenerator_CSSGeneration(t *testing.T) {
	input := `Style { css: ".title { color: red; }" } Window(id: "win") {}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, _ := parser.Parse()
	generator := NewGenerator()

	css := generator.GenerateCSS(program)
	if css != ".title { color: red; }\n" {
		t.Errorf("unexpected CSS output: %q", css)
	}
}

func TestGenerator_VBoxOrientation(t *testing.T) {
	input := `VBox(spacing: 5) {}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, _ := parser.Parse()
	generator := NewGenerator()
	output, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `<property name="orientation">vertical</property>`) {
		t.Error("VBox should have vertical orientation")
	}
}

func TestGenerator_MarginExpansion(t *testing.T) {
	input := `VBox(margin: 10) {}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, _ := parser.Parse()
	generator := NewGenerator()
	output, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	margins := []string{"margin-top", "margin-bottom", "margin-start", "margin-end"}
	for _, m := range margins {
		if !strings.Contains(output, `<property name="`+m+`">10</property>`) {
			t.Errorf("margin expansion missing: %s", m)
		}
	}
}

func TestGenerator_SortedProperties(t *testing.T) {
	input := `Window(title: "T", default_width: 800) {}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, _ := parser.Parse()
	generator := NewGenerator()
	output, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	posDefaultWidth := strings.Index(output, `name="default-width"`)
	posTitle := strings.Index(output, `name="title"`)
	if posDefaultWidth > posTitle {
		t.Error("properties should be sorted alphabetically")
	}
}

func TestGenerator_TranslatableMenu(t *testing.T) {
	input := `Menu(id: "app-menu") {
		Section {
			Item(label: "Open", action: "app.open")
				.translatable("label")
		}
	}
	Window() {}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	generator := NewGenerator()
	output, err := generator.Generate(program)
	if err != nil {
		t.Fatalf("unexpected generation error: %v", err)
	}

	expected := `<attribute name="label" translatable="yes">Open</attribute>`
	if !strings.Contains(output, expected) {
		t.Errorf("expected translatable attribute in menu item, got: %s", output)
	}
}

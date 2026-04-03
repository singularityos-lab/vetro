package vetro

import (
	"testing"
)

func TestParser_SimpleComponent(t *testing.T) {
	input := `Window(id: "main", title: "Test") {}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if program.Roots[0].Type != "Window" {
		t.Errorf("expected Window, got %s", program.Roots[0].Type)
	}
	if program.Roots[0].Properties["id"] != "main" {
		t.Errorf("expected id=main, got %v", program.Roots[0].Properties["id"])
	}
}

func TestParser_PositionalArgument(t *testing.T) {
	input := `Label("Hello World")`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if program.Roots[0].Properties["label"] != "Hello World" {
		t.Errorf("expected label=Hello World, got %v", program.Roots[0].Properties["label"])
	}
}

func TestParser_BooleanValues(t *testing.T) {
	input := `Window(resizable: false) {}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resizable, ok := program.Roots[0].Properties["resizable"].(bool)
	if !ok || resizable != false {
		t.Errorf("expected resizable=false (bool), got %v (%T)", program.Roots[0].Properties["resizable"], program.Roots[0].Properties["resizable"])
	}
}

func TestParser_IntegerValues(t *testing.T) {
	input := `Window(default_width: 800, default_height: 600) {}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	width, ok := program.Roots[0].Properties["default_width"].(int)
	if !ok || width != 800 {
		t.Errorf("expected default_width=800 (int), got %v (%T)", program.Roots[0].Properties["default_width"], program.Roots[0].Properties["default_width"])
	}
}

func TestParser_Modifiers(t *testing.T) {
	input := `Label("test").cssClass("title").halign("center")`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if program.Roots[0].Properties["css-classes"] != "title" {
		t.Errorf("expected css-classes=title, got %v", program.Roots[0].Properties["css-classes"])
	}
	if program.Roots[0].Properties["halign"] != "center" {
		t.Errorf("expected halign=center, got %v", program.Roots[0].Properties["halign"])
	}
}

func TestParser_Signals(t *testing.T) {
	input := `Button("Click me").onClicked(my_handler)`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if program.Roots[0].Signals["clicked"] != "my_handler" {
		t.Errorf("expected signal clicked=my_handler, got %v", program.Roots[0].Signals["clicked"])
	}
}

func TestParser_NestedComponents(t *testing.T) {
	input := `Window(id: "win") {
		VBox(spacing: 10) {
			Label("Hello")
			Button("World")
		}
	}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(program.Roots[0].Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(program.Roots[0].Children))
	}
	vbox := program.Roots[0].Children[0]
	if vbox.Type != "VBox" {
		t.Errorf("expected VBox, got %s", vbox.Type)
	}
	if len(vbox.Children) != 2 {
		t.Errorf("expected 2 children in VBox, got %d", len(vbox.Children))
	}
}

func TestParser_StyleBlock(t *testing.T) {
	input := `Style {
		css: ".title { font-size: 24px; }"
	}
	Window(id: "win") {}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(program.Styles) != 1 {
		t.Fatalf("expected 1 style block, got %d", len(program.Styles))
	}
	if program.Styles[0].CSS != ".title { font-size: 24px; }" {
		t.Errorf("unexpected CSS content: %s", program.Styles[0].CSS)
	}
}

func TestParser_ErrorPosition(t *testing.T) {
	input := `Window(}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	_, err := parser.Parse()
	if err == nil {
		t.Fatal("expected parse error")
	}
	if pErr, ok := err.(*ParserError); ok {
		if pErr.Token.Line == 0 {
			t.Error("expected error to include line number")
		}
	}
}

func TestParser_MarginExpansion(t *testing.T) {
	input := `VBox(margin: 24) {}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if program.Roots[0].Properties["margin"] != 24 {
		t.Errorf("expected margin=24, got %v", program.Roots[0].Properties["margin"])
	}
}

func TestParser_ChildType(t *testing.T) {
	input := `Window(id: "win") {
		HeaderBar().asChildType("titlebar")
		VBox(spacing: 10) {}
	}`
	lexer := NewLexer(input)
	tokens, _ := lexer.Tokenize()
	parser := NewParser(tokens)
	program, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(program.Roots[0].Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(program.Roots[0].Children))
	}
	headerBar := program.Roots[0].Children[0]
	if headerBar.ChildType != "titlebar" {
		t.Errorf("expected ChildType=titlebar, got %q", headerBar.ChildType)
	}
	vbox := program.Roots[0].Children[1]
	if vbox.ChildType != "" {
		t.Errorf("expected empty ChildType for VBox, got %q", vbox.ChildType)
	}
}

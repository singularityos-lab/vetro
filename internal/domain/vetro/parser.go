package vetro

import (
	"fmt"
	"strconv"
	"strings"
)

// Parser holds the tokens and current position.
type Parser struct {
	tokens []Token
	pos    int
}

// NewParser returns a new Parser instance.
func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

// current returns the current token without consuming it.
func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

// peek returns the next token without consuming.
func (p *Parser) peek() Token {
	if p.pos+1 >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos+1]
}

// advance consumes the current token and moves to the next.
func (p *Parser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

// consumeIf consumes the current token if it matches the
// expected type and returns true, otherwise returns false.
func (p *Parser) consumeIf(expected TokenType) bool {
	if p.current().Type == expected {
		p.advance()
		return true
	}
	return false
}

// expect consumes the current token if it matches, otherwise
// returns an error with position.
func (p *Parser) expect(expected TokenType) error {
	tok := p.current()
	if tok.Type != expected {
		return newParseError(
			tok,
			"expected %s, got %s ('%s')",
			tokenTypeName(expected),
			tokenTypeName(tok.Type),
			tok.Literal,
		)
	}
	p.advance()
	return nil
}

// tokenTypeName returns a human-readable name for a token type.
func tokenTypeName(t TokenType) string {
	switch t {
	case TokenEOF:
		return "end of file"
	case TokenIdent:
		return "identifier"
	case TokenString:
		return "string"
	case TokenNumber:
		return "number"
	case TokenLParen:
		return "'('"
	case TokenRParen:
		return "')'"
	case TokenLBrace:
		return "'{'"
	case TokenRBrace:
		return "'}'"
	case TokenColon:
		return "':'"
	case TokenDot:
		return "'.'"
	case TokenComma:
		return "','"
	default:
		return "unknown"
	}
}

// Parse parses the tokens and returns a Program containing the root components and styles.
func (p *Parser) Parse() (*Program, error) {
	var styles []StyleBlock
	var menus []*ComponentNode
	var roots []*ComponentNode
	var requires *Requires

	// Parse style blocks, menu blocks, requires, and top-level components
	for p.current().Type != TokenEOF {
		if p.current().Type == TokenIdent && p.current().Literal == "Requires" {
			// Parse Requires(lib: "...", version: "...")
			req, err := p.parseRequires()
			if err != nil {
				return nil, err
			}
			requires = req
		} else if p.current().Type == TokenIdent && p.current().Literal == "Style" {
			// Parse a Style block
			style, err := p.parseStyleBlock()
			if err != nil {
				return nil, err
			}
			styles = append(styles, *style)
		} else if p.current().Type == TokenIdent && p.current().Literal == "Menu" {
			// Parse a Menu block (top-level GMenu definition)
			menu, err := p.parseComponent()
			if err != nil {
				return nil, err
			}
			menus = append(menus, menu)
		} else if p.current().Type == TokenIdent {
			node, err := p.parseComponent()
			if err != nil {
				return nil, err
			}
			roots = append(roots, node)
		} else {
			return nil, newParseError(p.current(), "unexpected token '%s', expected 'Requires', 'Style', 'Menu', or component type", p.current().Literal)
		}
	}

	if len(roots) == 0 {
		return nil, fmt.Errorf("no root component found in file")
	}

	return &Program{
		Roots:    roots,
		Styles:   styles,
		Menus:    menus,
		Requires: requires,
	}, nil
}

// parseStyleBlock parses a Style { css: "..." } block.
// parseRequires parses a Requires(lib: "...", version: "...") block.
func (p *Parser) parseRequires() (*Requires, error) {
	p.advance()
	// Expect (
	if err := p.expect(TokenLParen); err != nil {
		return nil, err
	}

	req := &Requires{}
	for p.current().Type != TokenRParen && p.current().Type != TokenEOF {
		if p.current().Type == TokenIdent {
			key := p.current().Literal
			p.advance()
			if err := p.expect(TokenColon); err != nil {
				return nil, err
			}
			value, err := p.parseValue()
			if err != nil {
				return nil, err
			}
			strVal := fmt.Sprintf("%v", value)
			switch key {
			case "lib":
				req.Lib = strVal
			case "version":
				req.Version = strVal
			default:
				return nil, newParseError(p.current(), "unknown Requires property '%s'", key)
			}
			// Optional comma
			p.consumeIf(TokenComma)
		} else {
			break
		}
	}
	// Expect )
	if err := p.expect(TokenRParen); err != nil {
		return nil, err
	}
	return req, nil
}

func (p *Parser) parseStyleBlock() (*StyleBlock, error) {
	p.advance()
	// Expect {
	if err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	css := ""
	priority := "application"

	for {
		if p.current().Type == TokenEOF {
			return nil, newParseError(p.current(), "unexpected end of file inside Style block")
		}
		if p.current().Type == TokenRBrace {
			p.advance()
			break
		}
		if p.current().Type == TokenIdent {
			key := p.current().Literal
			p.advance()
			if err := p.expect(TokenColon); err != nil {
				return nil, err
			}
			value, err := p.parseValue()
			if err != nil {
				return nil, err
			}
			switch key {
			case "css":
				css = fmt.Sprintf("%v", value)
			case "priority":
				priority = fmt.Sprintf("%v", value)
			default:
				return nil, newParseError(p.current(), "unknown Style property '%s'", key)
			}
			// Optional comma
			p.consumeIf(TokenComma)
		} else {
			return nil, newParseError(p.current(), "unexpected token '%s' in Style block", p.current().Literal)
		}
	}

	return &StyleBlock{
		CSS:      css,
		Priority: priority,
	}, nil
}

func (p *Parser) parseComponent() (*ComponentNode, error) {
	tok := p.current()
	if tok.Type != TokenIdent {
		return nil, newParseError(tok, "expected component type (e.g., Window, VBox, Label), got '%s'", tok.Literal)
	}
	compType := tok.Literal
	p.advance()

	// Support dotted namespace: Singularity.TabBar → stored as "Singularity.TabBar"
	// and resolved to the full C type (e.g. SingularityTabBar) by qualifyGtkClassName.
	for p.current().Type == TokenDot {
		// Peek: if next-next is an ident followed by '(' or '.', it's a namespace qualifier.
		// If next-next is a modifier keyword (known modifier), it's a method — stop.
		if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == TokenIdent {
			nextLit := p.tokens[p.pos+1].Literal
			// A namespace segment starts with uppercase
			if len(nextLit) > 0 && nextLit[0] >= 'A' && nextLit[0] <= 'Z' {
				p.advance() // consume '.'
				sub := p.current()
				compType += "." + sub.Literal
				p.advance()
				continue
			}
		}
		break
	}

	var args map[string]any
	if p.consumeIf(TokenLParen) {
		var err error
		args, err = p.parseArguments()
		if err != nil {
			return nil, err
		}
		if err := p.expect(TokenRParen); err != nil {
			return nil, err
		}
	}

	// Parse zero or more modifiers
	signals := make(map[string]string)
	signalAttrs := make(map[string]SignalAttr)
	childType := ""           // for special child types like "titlebar"
	childName := ""           // for Stack page names
	var translatable []string // property names marked as translatable
	for p.consumeIf(TokenDot) {
		tok := p.current()
		if tok.Type != TokenIdent {
			return nil, newParseError(tok, "expected modifier name after '.', got '%s'", tok.Literal)
		}
		modName := tok.Literal
		modTok := tok // save for error messages
		p.advance()
		if err := p.expect(TokenLParen); err != nil {
			return nil, err
		}
		modArgs, err := p.parseModifierArgumentList()
		if err != nil {
			return nil, err
		}
		if err := p.expect(TokenRParen); err != nil {
			return nil, err
		}

		// Handle translatable modifier (takes multiple string args)
		if modName == "translatable" {
			for _, arg := range modArgs {
				translatable = append(translatable, fmt.Sprintf("%v", arg))
			}
		} else {
			// All other modifiers take exactly one argument
			if len(modArgs) != 1 {
				return nil, newParseError(modTok, "modifier '%s' expects exactly one argument", modName)
			}

			// Handle asChildType modifier
			if modName == "asChildType" {
				childType = fmt.Sprintf("%v", modArgs[0])
			} else if modName == "asChildName" {
				// For Stack pages: .asChildName("page1")
				childName = fmt.Sprintf("%v", modArgs[0])
			} else if modName == "parentType" {
				// Store as parentType (special property for templates)
				if args == nil {
					args = make(map[string]any)
				}
				args["parentType"] = modArgs[0]
			} else if modName == "swapped" || modName == "after" {
				// Signal attributes: .swapped("yes") or .after("yes")
				val := fmt.Sprintf("%v", modArgs[0])
				if len(signals) > 0 {
					var lastSig string
					for k := range signals {
						lastSig = k
					}
					attr := signalAttrs[lastSig]
					if modName == "swapped" {
						attr.Swapped = (val == "yes" || val == "true")
					} else {
						attr.After = (val == "yes" || val == "true")
					}
					signalAttrs[lastSig] = attr
				}
			} else if strings.HasPrefix(modName, "on") && len(modName) > 2 {
				// Signal handler (e.g., onClicked -> clicked)
				signalName := strings.ToLower(modName[2:3]) + modName[3:]
				signalName = toKebabCase(signalName)

				// Validate signal name for this widget type
				gtkClass := vetroToGtkClass(compType)
				if validSigs := LookupValidSignals(gtkClass); validSigs != nil {
					found := false
					for _, vs := range validSigs {
						if vs == signalName {
							found = true
							break
						}
					}
					if !found {
						return nil, newParseError(modTok, "signal '%s' is not valid for %s", signalName, gtkClass)
					}
				}

				signals[signalName] = fmt.Sprintf("%v", modArgs[0])
			} else {
				// Regular property modifier
				propertyName := toKebabCase(modName)
				if args == nil {
					args = make(map[string]any)
				}
				// css-classes accumulates across multiple .cssClass() calls
				if propertyName == "css-classes" {
					if existing, ok := args[propertyName]; ok {
						args[propertyName] = fmt.Sprintf("%v %v", existing, modArgs[0])
					} else {
						args[propertyName] = modArgs[0]
					}
				} else {
					args[propertyName] = modArgs[0]
				}
			}
		}
	}

	// Expect opening brace for the body
	var hasBody bool
	if p.consumeIf(TokenLBrace) {
		hasBody = true
	} else {
		hasBody = false
	}

	// Parse properties and children
	props := make(map[string]any)
	var children []*ComponentNode
	if hasBody {
		for {
			tok := p.current()
			if tok.Type == TokenEOF {
				return nil, newParseError(tok, "unexpected end of file inside component body, did you forget '}'?")
			}
			if tok.Type == TokenRBrace {
				p.advance() // consume '}'
				break
			}

			// Check if it's a property assignment (IDENT ':')
			if tok.Type == TokenIdent {
				// Look ahead to see if it's a property or a child component
				savePos := p.pos
				ident := tok.Literal
				p.advance()
				if p.consumeIf(TokenColon) {
					// It's a property
					value, err := p.parseValue()
					if err != nil {
						return nil, err
					}
					props[ident] = value
				} else {
					// It's a child component: backtrack and parse as component
					p.pos = savePos
					child, err := p.parseComponent()
					if err != nil {
						return nil, err
					}
					children = append(children, child)
				}
			} else {
				return nil, newParseError(tok, "unexpected token '%s' in component body", tok.Literal)
			}
		}
	}

	// Merge args into props
	for k, v := range args {
		props[k] = v
	}

	// Validate and convert property types for this component
	convertedProps := make(map[string]any)
	for k, v := range props {
		schema := LookupPropertySchema(k)
		if schema != nil {
			converted, err := convertPropertyType(k, v, *schema)
			if err != nil {
				return nil, newParseError(tok, "property '%s': %v", k, err)
			}
			convertedProps[k] = converted
		} else {
			// Unknown property
			convertedProps[k] = v
		}
	}

	return &ComponentNode{
		Type:         compType,
		Properties:   convertedProps,
		Signals:      signals,
		SignalAttrs:  signalAttrs,
		ChildType:    childType,
		ChildName:    childName,
		Translatable: translatable,
		Children:     children,
	}, nil
}

// parseArguments parses either a single positional argument or key-value pairs.
func (p *Parser) parseArguments() (map[string]any, error) {
	args := make(map[string]any)
	if p.current().Type == TokenRParen {
		return args, nil
	}

	// Check if first token is string, number, or boolean
	if p.current().Type == TokenString || p.current().Type == TokenNumber ||
		(p.current().Type == TokenIdent && (p.current().Literal == "true" || p.current().Literal == "false")) {
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		args["label"] = value
		return args, nil
	}

	// Parse key-value pairs
	for {
		if p.current().Type == TokenEOF {
			return nil, newParseError(p.current(), "unexpected end of file in argument list")
		}
		if p.current().Type == TokenRParen {
			break
		}
		// Parse key
		keyTok := p.current()
		if keyTok.Type != TokenIdent {
			return nil, newParseError(keyTok, "expected property name, got '%s'", keyTok.Literal)
		}
		key := keyTok.Literal
		p.advance()
		// Expect colon
		if err := p.expect(TokenColon); err != nil {
			return nil, err
		}
		// Parse value
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		args[key] = value

		// Check for comma or closing paren
		if p.consumeIf(TokenComma) {
			continue
		}
		if p.current().Type == TokenRParen {
			break
		}
		return nil, newParseError(p.current(), "expected ',' or ')', got '%s'", p.current().Literal)
	}
	return args, nil
}

// parseModifierArgumentList parses a comma-separated list of values.
func (p *Parser) parseModifierArgumentList() ([]any, error) {
	var args []any
	for {
		if p.current().Type == TokenEOF {
			return nil, newParseError(p.current(), "unexpected end of file in modifier arguments")
		}
		if p.current().Type == TokenRParen {
			break
		}
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		args = append(args, value)
		if p.consumeIf(TokenComma) {
			continue
		}
		if p.current().Type == TokenRParen {
			break
		}
		return nil, newParseError(p.current(), "expected ',' or ')', got '%s'", p.current().Literal)
	}
	return args, nil
}

// parseValue parses a value: string, number, boolean, or identifier.
func (p *Parser) parseValue() (any, error) {
	tok := p.current()
	switch tok.Type {
	case TokenString:
		p.advance()
		return tok.Literal, nil
	case TokenNumber:
		p.advance()
		// Try to parse as integer
		if val, err := strconv.Atoi(tok.Literal); err == nil {
			return val, nil
		}
		return tok.Literal, nil
	case TokenIdent:
		p.advance()
		// Handle boolean keywords
		if tok.Literal == "true" {
			return true, nil
		}
		if tok.Literal == "false" {
			return false, nil
		}
		return tok.Literal, nil
	default:
		return nil, newParseError(tok, "expected value (string, number, boolean, or identifier), got '%s'", tok.Literal)
	}
}

// vetroToGtkClass converts a Vetro component type to GTK class name.
func vetroToGtkClass(componentType string) string {
	switch componentType {
	case "VBox", "HBox", "Box":
		return "GtkBox"
	default:
		// Dotted namespace: "Singularity.TabBar" → search in metadata
		if strings.Contains(componentType, ".") {
			parts := strings.SplitN(componentType, ".", 2)
			ns, simple := parts[0], parts[1]
			// Try exact prefixed form first
			full := ns + simple
			if MetadataManager != nil && MetadataManager.Metadata != nil {
				if _, ok := MetadataManager.Metadata.Classes[full]; ok {
					return full
				}
			}
			return full
		}
		return qualifyGtkClassName(componentType)
	}
}

// convertPropertyType converts and validates a property value to the expected GTK type.
func convertPropertyType(name string, value any, schema PropertySchema) (any, error) {
	switch schema.Type {
	case PropTypeBool:
		switch v := value.(type) {
		case bool:
			return v, nil
		case string:
			if v == "true" {
				return true, nil
			}
			if v == "false" {
				return false, nil
			}
			return nil, fmt.Errorf("expected boolean value (true/false), got '%s'", v)
		default:
			return nil, fmt.Errorf("expected boolean value, got %T", value)
		}
	case PropTypeInt:
		switch v := value.(type) {
		case int:
			return v, nil
		case string:
			if val, err := strconv.Atoi(v); err == nil {
				return val, nil
			}
			return v, nil // Keep as string for XML
		default:
			return value, nil
		}
	case PropTypeEnum:
		strVal := fmt.Sprintf("%v", value)
		if len(schema.EnumVals) > 0 {
			found := false
			for _, ev := range schema.EnumVals {
				if ev == strVal {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("invalid enum value '%s' for property '%s', valid values: %v", strVal, name, schema.EnumVals)
			}
		}
		return strVal, nil
	case PropTypeString:
		return fmt.Sprintf("%v", value), nil
	default:
		return value, nil
	}
}

// toKebabCase converts a camelCase string to kebab-case.
func toKebabCase(s string) string {
	if s == "cssClass" || s == "cssClasses" || s == "css_classes" {
		return "css-classes"
	}
	if s == "css_class" {
		return "css-class"
	}
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '-', r+32)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

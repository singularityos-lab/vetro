package vetro

import (
	"fmt"
	"strings"
)

// ParserError represents a parse error with source position and suggestions.
type ParserError struct {
	Message    string
	Token      Token
	Suggestion string
}

func (e *ParserError) Error() string {
	msg := fmt.Sprintf(":%d:%d: %s", e.Token.Line, e.Token.Col, e.Message)
	if e.Suggestion != "" {
		msg += "\n  -> " + e.Suggestion
	}
	return msg
}

// newParseError creates a ParserError with a formatted message.
func newParseError(tok Token, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	suggestion := getSuggestion(msg)
	return &ParserError{
		Message:    msg,
		Token:      tok,
		Suggestion: suggestion,
	}
}

// getSuggestion returns a helpful hint based on the error message.
func getSuggestion(msg string) string {
	switch {
	case strings.Contains(msg, "expected ')'"):
		return "check that every '(' has a matching ')'"
	case strings.Contains(msg, "expected '}'"):
		return "check that every '{' has a matching '}'"
	case strings.Contains(msg, "expected component type"):
		return "start with Window, VBox, Label, Button, or another widget"
	case strings.Contains(msg, "expected ':'"):
		return "properties use 'name: value' syntax"
	case strings.Contains(msg, "expected value"):
		return "provide a value: \"text\", 42, true, or false"
	case strings.Contains(msg, "expected ','"):
		return "separate arguments with commas"
	case strings.Contains(msg, "unexpected token"):
		return "check for missing braces or typos"
	case strings.Contains(msg, "unterminated"):
		return "make sure every \" has a matching closing quote"
	case strings.Contains(msg, "only one root"):
		return "wrap everything in a single Window or ApplicationWindow"
	case strings.Contains(msg, "no root component"):
		return "add a Window, ApplicationWindow, or other container as the root"
	case strings.Contains(msg, "signal"):
		return "check GTK docs for valid signals for this widget"
	case strings.Contains(msg, "expects exactly one argument"):
		return "provide exactly one argument in parentheses"
	default:
		return ""
	}
}

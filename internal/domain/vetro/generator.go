package vetro

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"
)

// Generator generates XML from the AST.
type Generator struct{}

// NewGenerator returns a new Generator instance.
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate converts the Program AST to XML string.
func (g *Generator) Generate(program *Program) (string, error) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString("<interface>" + "\n")

	// Emit requires if present
	if program.Requires != nil {
		fmt.Fprintf(&b, "  <requires lib=\"%s\" version=\"%s\"/>\n", program.Requires.Lib, program.Requires.Version)
	}

	// Generate CSS styles first
	for _, style := range program.Styles {
		g.generateStyle(&b, style, 2)
	}

	// Generate menu definitions
	for _, menu := range program.Menus {
		if err := g.generateNode(&b, menu, 2); err != nil {
			return "", err
		}
	}

	// Generate all top-level components
	for _, root := range program.Roots {
		if err := g.generateNode(&b, root, 2); err != nil {
			return "", err
		}
	}

	b.WriteString("</interface>\n")
	return b.String(), nil
}

// GenerateCSS generates a separate CSS file from Style blocks.
func (g *Generator) GenerateCSS(program *Program) string {
	var css strings.Builder
	for _, style := range program.Styles {
		if style.CSS != "" {
			css.WriteString(style.CSS)
			css.WriteString("\n")
		}
	}
	return css.String()
}

// generateStyle writes a CSS style block as a custom XML element.
func (g *Generator) generateStyle(b *strings.Builder, style StyleBlock, indent int) {
	indentStr := strings.Repeat(" ", indent)
	fmt.Fprintf(b, "%s<!-- CSS Style Block (priority: %s) -->\n", indentStr, style.Priority)
	fmt.Fprintf(b, "%s<custom>\n", indentStr)
	escapedCSS := g.xmlEscape(style.CSS)
	fmt.Fprintf(b, "%s  <css>%s</css>\n", indentStr, escapedCSS)
	fmt.Fprintf(b, "%s</custom>\n", indentStr)
}

// getGtkClass returns the GTK class name for a Vetro component type.
func (g *Generator) getGtkClass(componentType string) string {
	switch componentType {
	case "VBox", "HBox", "Box":
		return "GtkBox"
	default:
		return vetroToGtkClass(componentType)
	}
}

// convertPropertyName converts a Vetro property name to GTK property name.
func (g *Generator) convertPropertyName(key string) string {
	switch key {
	case "cssClass", "cssClasses", "css_classes":
		return "css-classes"
	case "css_class":
		return "css-class"
	case "parentType":
		return "parent-type"
	}
	return camelToKebab(key)
}

// camelToKebab converts a camelCase or snake_case string to kebab-case.
func camelToKebab(s string) string {
	var result []rune
	for i, r := range s {
		if r == '_' {
			result = append(result, '-')
		} else if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '-', r+32) // convert to lowercase
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// formatValue formats a value for XML output.
func (g *Generator) formatValue(value any) string {
	switch v := value.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int:
		return fmt.Sprintf("%d", v)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// xmlEscape escapes special XML characters in a string.
func (g *Generator) xmlEscape(s string) string {
	var buf bytes.Buffer
	xml.Escape(&buf, []byte(s))
	return buf.String()
}

func (g *Generator) generateNode(b *strings.Builder, node *ComponentNode, indent int) error {
	indentStr := strings.Repeat(" ", indent)
	class := g.getGtkClass(node.Type)

	props := make(map[string]any)
	maps.Copy(props, node.Properties)

	if node.Type == "Menu" {
		menuID := ""
		if idVal, ok := props["id"]; ok {
			menuID = fmt.Sprintf("%v", idVal)
		}
		fmt.Fprintf(b, "%s<menu id=\"%s\">\n", indentStr, menuID)
		for _, child := range node.Children {
			g.generateMenuElement(b, child, indent+2)
		}
		fmt.Fprintf(b, "%s</menu>\n", indentStr)
		return nil
	}

	// Template: generates <template class="..." parent="..."> with children
	// as direct <object> elements (no <child> wrapper) — for GtkTemplate support.
	if node.Type == "Template" {
		templateClass := ""
		templateParent := ""
		if v, ok := props["class"]; ok {
			delete(props, "class")
			templateClass = fmt.Sprintf("%v", v)
		}
		if v, ok := props["parent"]; ok {
			delete(props, "parent")
			templateParent = qualifyGtkClassName(fmt.Sprintf("%v", v))
		}
		fmt.Fprintf(b, "%s<template class=\"%s\" parent=\"%s\">\n", indentStr, templateClass, templateParent)
		for _, child := range node.Children {
			if err := g.generateNode(b, child, indent+2); err != nil {
				return err
			}
		}
		fmt.Fprintf(b, "%s</template>\n", indentStr)
		return nil
	}

	// Extract id if present
	idAttr := ""
	if idVal, ok := props["id"]; ok {
		delete(props, "id")
		idAttr = fmt.Sprintf(` id="%s"`, g.xmlEscape(g.formatValue(idVal)))
	}

	// Extract parentType for templates
	parentType := ""
	if pt, ok := props["parentType"]; ok {
		delete(props, "parentType")
		parentType = fmt.Sprintf("%v", pt)
		parentType = qualifyGtkClassName(parentType)
	}

	// Write the object/template opening tag
	if parentType != "" {
		fmt.Fprintf(b, "%s<template class=\"%s\" parent=\"%s\"%s>\n", indentStr, node.Type, parentType, idAttr)
	} else {
		fmt.Fprintf(b, "%s<object class=\"%s\"%s>\n", indentStr, class, idAttr)
	}

	// Handle special cases based on component type
	switch node.Type {
	case "VBox", "HBox":
		// Only set orientation if not already specified
		if _, ok := props["orientation"]; !ok {
			if node.Type == "VBox" {
				props["orientation"] = "vertical"
			} else {
				props["orientation"] = "horizontal"
			}
		}
	}

	// Handle margin expansion
	if marginVal, ok := props["margin"]; ok {
		delete(props, "margin")
		props["margin-top"] = marginVal
		props["margin-bottom"] = marginVal
		props["margin-start"] = marginVal
		props["margin-end"] = marginVal
	}

	// Write properties sorted by key
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := props[key]
		gtkKey := g.convertPropertyName(key)
		formattedValue := g.formatValue(value)
		escapedValue := g.xmlEscape(formattedValue)

		// Check if this poperty is translatable
		translatableAttr := ""
		for _, tp := range node.Translatable {
			if tp == key || gtkToVetroProperty(tp) == key {
				translatableAttr = ` translatable="yes"`
				break
			}
		}

		fmt.Fprintf(b, "%s  <property name=\"%s\"%s>%s</property>\n", indentStr, gtkKey, translatableAttr, escapedValue)
	}

	// Write signals sorted by key
	sigKeys := make([]string, 0, len(node.Signals))
	for k := range node.Signals {
		sigKeys = append(sigKeys, k)
	}
	sort.Strings(sigKeys)

	for _, key := range sigKeys {
		handler := node.Signals[key]
		escapedKey := g.xmlEscape(key)
		escapedHandler := g.xmlEscape(handler)

		// Check for signal attributes
		swapped := "no"
		after := ""
		if attrs, ok := node.SignalAttrs[key]; ok {
			if attrs.Swapped {
				swapped = "yes"
			}
			if attrs.After {
				after = ` after="yes"`
			}
		}

		fmt.Fprintf(b, "%s  <signal name=\"%s\" handler=\"%s\" swapped=\"%s\"%s/>\n", indentStr, escapedKey, escapedHandler, swapped, after)
	}

	// Write children
	for _, child := range node.Children {
		childAttrs := ""
		if child.ChildType != "" {
			childAttrs += fmt.Sprintf(` type="%s"`, g.xmlEscape(child.ChildType))
		}
		if child.ChildName != "" {
			childAttrs += fmt.Sprintf(` name="%s"`, g.xmlEscape(child.ChildName))
		}
		fmt.Fprintf(b, "%s  <child%s>\n", indentStr, childAttrs)
		if err := g.generateNode(b, child, indent+4); err != nil {
			return err
		}
		fmt.Fprintf(b, "%s  </child>\n", indentStr)
	}

	// Write the closing tag
	if parentType != "" {
		fmt.Fprintf(b, "%s</template>\n", indentStr)
	} else {
		fmt.Fprintf(b, "%s</object>\n", indentStr)
	}

	return nil
}

// generateMenuElement emits a GMenu child element (section or item).
func (g *Generator) generateMenuElement(b *strings.Builder, node *ComponentNode, indent int) {
	indentStr := strings.Repeat(" ", indent)

	switch node.Type {
	case "Section":
		fmt.Fprintf(b, "%s<section>\n", indentStr)
		for _, child := range node.Children {
			g.generateMenuElement(b, child, indent+2)
		}
		fmt.Fprintf(b, "%s</section>\n", indentStr)

	case "Item":
		fmt.Fprintf(b, "%s<item>\n", indentStr)
		// Write attributes
		keys := make([]string, 0, len(node.Properties))
		for k := range node.Properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := node.Properties[key]
			escapedKey := g.xmlEscape(key)
			escapedValue := g.xmlEscape(g.formatValue(value))

			translatable := slices.Contains(node.Translatable, key)
			translatableAttr := ""
			if translatable {
				translatableAttr = ` translatable="yes"`
			}
			fmt.Fprintf(b, "%s  <attribute name=\"%s\"%s>%s</attribute>\n", indentStr, escapedKey, translatableAttr, escapedValue)
		}
		fmt.Fprintf(b, "%s</item>\n", indentStr)

	default:
		// Recursively handle any other nested elements
		for _, child := range node.Children {
			g.generateMenuElement(b, child, indent)
		}
	}
}

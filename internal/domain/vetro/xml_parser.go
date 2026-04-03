package vetro

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
)

// --- XML DOM structures for parsing GTK Builder ---

type xmlInterface struct {
	XMLName   xml.Name     `xml:"interface"`
	Requires  []xmlRequire `xml:"requires"`
	Templates []xmlObject  `xml:"template"`
	Objects   []xmlObject  `xml:"object"`
	Customs   []xmlCustom  `xml:"custom"`
	Menus     []xmlMenu    `xml:"menu"`
}

type xmlRequire struct {
	Lib     string `xml:"lib,attr"`
	Version string `xml:"version,attr"`
}

type xmlCustom struct {
	CSS string `xml:"css"`
}

type xmlMenu struct {
	ID       string           `xml:"id,attr"`
	Sections []xmlMenuSection `xml:"section"`
}

type xmlMenuSection struct {
	Items []xmlMenuItem `xml:"item"`
}

type xmlMenuItem struct {
	Attributes []xmlMenuItemAttr `xml:"attribute"`
}

type xmlMenuItemAttr struct {
	Name         string `xml:"name,attr"`
	Translatable string `xml:"translatable,attr"`
	Value        string `xml:",chardata"`
}

type xmlObject struct {
	Class      string        `xml:"class,attr"`
	ID         string        `xml:"id,attr"`
	Parent     string        `xml:"parent,attr"`
	Properties []xmlProperty `xml:"property"`
	Signals    []xmlSignal   `xml:"signal"`
	Children   []xmlChild    `xml:"child"`
	Styles     []xmlStyle    `xml:"style"`
}

type xmlProperty struct {
	Name         string     `xml:"name,attr"`
	Translatable string     `xml:"translatable,attr"`
	Value        string     `xml:",chardata"`
	Object       *xmlObject `xml:"object"`
}

type xmlSignal struct {
	Name    string `xml:"name,attr"`
	Handler string `xml:"handler,attr"`
}

type xmlChild struct {
	Type   string    `xml:"type,attr"`
	Name   string    `xml:"name,attr"`
	Object xmlObject `xml:"object"`
}

type xmlStyle struct {
	Classes []xmlStyleClass `xml:"class"`
}

type xmlStyleClass struct {
	Name string `xml:"name,attr"`
}

// VetroEmitter converts GTK XML back to Vetro source code.
type VetroEmitter struct{}

// NewVetroEmitter creates a new VetroEmitter.
func NewVetroEmitter() *VetroEmitter {
	return &VetroEmitter{}
}

// EmitXMLToVetro parses GTK XML and returns Vetro source code.
func (e *VetroEmitter) EmitXMLToVetro(xmlSource string) (string, error) {
	var iface xmlInterface
	if err := xml.Unmarshal([]byte(xmlSource), &iface); err != nil {
		return "", fmt.Errorf("failed to parse XML: %w", err)
	}

	var b strings.Builder

	// Emit requires as Vetro syntax if present
	for _, req := range iface.Requires {
		fmt.Fprintf(&b, "Requires(lib: %q, version: %q)\n", req.Lib, req.Version)
	}

	// Emit style blocks if present
	for _, custom := range iface.Customs {
		if custom.CSS != "" {
			b.WriteString("Style {\n")
			b.WriteString(fmt.Sprintf("    css: \"%s\"\n", escapeVetroString(custom.CSS)))
			b.WriteString("}\n\n")
		}
	}

	// Emit menu definitions declaratively
	for _, menu := range iface.Menus {
		if err := e.emitMenu(&b, menu, 0); err != nil {
			return "", err
		}
		b.WriteString("\n")
	}

	// Find the root widget: prefer <template>, then <object>
	var rootObj *xmlObject
	if len(iface.Templates) > 0 {
		rootObj = &iface.Templates[0]
	} else if len(iface.Objects) > 0 {
		rootObj = &iface.Objects[0]
	}

	if rootObj != nil {
		if err := e.emitObject(&b, *rootObj, 0, "", ""); err != nil {
			return "", err
		}
	}

	return b.String(), nil
}

func (e *VetroEmitter) emitObject(b *strings.Builder, obj xmlObject, indent int, childType string, childName string) error {
	indentStr := strings.Repeat("    ", indent)

	// Convert GTK class to Vetro type
	vetroType := gtkClassToVetroType(obj.Class)

	// Collect arguments (id + key-value properties)
	var args []string
	if obj.ID != "" {
		args = append(args, fmt.Sprintf("id: %q", obj.ID))
	}

	// Build a map of properties
	propMap := make(map[string]string)
	for _, p := range obj.Properties {
		if p.Object != nil {
			continue
		}
		propMap[p.Name] = normalizeValue(p.Value)
	}

	// Handle margin compression
	marginVal := ""
	marginCount := 0
	if v, ok := propMap["margin-top"]; ok {
		marginVal = v
		marginCount++
	}
	if v, ok := propMap["margin-bottom"]; ok && v == marginVal {
		marginCount++
	}
	if v, ok := propMap["margin-start"]; ok && v == marginVal {
		marginCount++
	}
	if v, ok := propMap["margin-end"]; ok && v == marginVal {
		marginCount++
	}
	if marginCount == 4 && marginVal != "" {
		args = append(args, fmt.Sprintf("margin: %s", formatVetroValue(marginVal)))
		delete(propMap, "margin-top")
		delete(propMap, "margin-bottom")
		delete(propMap, "margin-start")
		delete(propMap, "margin-end")
	}

	// GtkBox -> VBox/HBox based on orientation
	if obj.Class == "GtkBox" {
		if v, ok := propMap["orientation"]; ok {
			delete(propMap, "orientation")
			if v == "vertical" {
				vetroType = "VBox"
			} else {
				vetroType = "HBox"
			}
		}
	}

	// Properties that become modifiers
	modifierProps := map[string]bool{
		"css-classes": true,
		"halign":      true,
		"valign":      true,
	}

	// Collecting key-value properties
	var kvProps []string
	sortedKeys := make([]string, 0, len(propMap))
	for k := range propMap {
		if !modifierProps[k] {
			sortedKeys = append(sortedKeys, k)
		}
	}
	sort.Strings(sortedKeys)
	for _, k := range sortedKeys {
		vetroKey := gtkToVetroProperty(k)
		kvProps = append(kvProps, fmt.Sprintf("%s: %s", vetroKey, formatVetroValue(propMap[k])))
	}
	args = append(args, kvProps...)

	// Build component line
	if len(args) > 0 {
		fmt.Fprintf(b, "%s%s(%s)\n", indentStr, vetroType, strings.Join(args, ", "))
	} else {
		fmt.Fprintf(b, "%s%s()\n", indentStr, vetroType)
	}

	// Emit parent type for templates (custom widget subclassing)
	if obj.Parent != "" {
		parentType := gtkClassToVetroType(obj.Parent)
		fmt.Fprintf(b, "%s    .parentType(%q)\n", indentStr, parentType)
	}

	// Emit translatable modifier for properties with translatable="yes"
	var translatableProps []string
	for _, p := range obj.Properties {
		if p.Translatable == "yes" && p.Object == nil {
			translatableProps = append(translatableProps, gtkToVetroProperty(p.Name))
		}
	}

	if len(translatableProps) > 0 {
		args := make([]string, len(translatableProps))
		for i, p := range translatableProps {
			args[i] = fmt.Sprintf("%q", p)
		}
		fmt.Fprintf(b, "%s    .translatable(%s)\n", indentStr, strings.Join(args, ", "))
	}

	// Emit cssClass from css-classes property
	if cssClasses, ok := propMap["css-classes"]; ok {
		fmt.Fprintf(b, "%s    .cssClass(%q)\n", indentStr, cssClasses)
	}

	// Emit cssClass from <style><class> elements
	for _, style := range obj.Styles {
		for _, cls := range style.Classes {
			fmt.Fprintf(b, "%s    .cssClass(%q)\n", indentStr, cls.Name)
		}
	}

	// Emit halign/valign modifiers
	for _, align := range []string{"halign", "valign"} {
		if v, ok := propMap[align]; ok {
			fmt.Fprintf(b, "%s    .%s(%q)\n", indentStr, align, v)
		}
	}

	// Emit asChildType
	if childType != "" {
		fmt.Fprintf(b, "%s    .asChildType(%q)\n", indentStr, childType)
	}

	// Emit asChildName (for Stack pages)
	if childName != "" {
		fmt.Fprintf(b, "%s    .asChildName(%q)\n", indentStr, childName)
	}

	// Emit signals
	for _, sig := range obj.Signals {
		handlerName := sig.Handler
		signalName := kebabToCamelCase(sig.Name)
		fmt.Fprintf(b, "%s    .on%s(%s)\n", indentStr, capitalize(signalName), handlerName)
	}

	// Collect all children
	hasChildren := len(obj.Children) > 0
	for _, p := range obj.Properties {
		if p.Object != nil {
			hasChildren = true
			break
		}
	}

	if hasChildren {
		fmt.Fprintf(b, "%s{\n", indentStr)

		// Emit inline children (from properties like titlebar, child)
		for _, p := range obj.Properties {
			if p.Object != nil {
				inlineChildType := ""
				if p.Name == "titlebar" {
					inlineChildType = "titlebar"
				}
				if err := e.emitObject(b, *p.Object, indent+1, inlineChildType, ""); err != nil {
					return err
				}
			}
		}

		// Emit regular children
		for _, child := range obj.Children {
			if err := e.emitObject(b, child.Object, indent+1, child.Type, child.Name); err != nil {
				return err
			}
		}

		fmt.Fprintf(b, "%s}\n", indentStr)
	}

	return nil
}

// normalizeValue normalizes boolean strings (True/true -> true, False/false -> false).
func normalizeValue(v string) string {
	switch v {
	case "True", "TRUE":
		return "true"
	case "False", "FALSE":
		return "false"
	default:
		return v
	}
}

// gtkClassToVetroType converts a GTK class name to a Vetro type name.
func gtkClassToVetroType(class string) string {
	if strings.HasPrefix(class, "Gtk") {
		return class[3:]
	}
	return class
}

// gtkToVetroProperty converts a GTK property name (kebab-case) to Vetro (camelCase).
func gtkToVetroProperty(prop string) string {
	parts := strings.Split(prop, "-")
	if len(parts) == 1 {
		return prop
	}
	result := parts[0]
	for _, p := range parts[1:] {
		if len(p) > 0 {
			result += capitalize(p)
		}
	}
	return result
}

// kebabToCamelCase converts kebab-case to camelCase.
func kebabToCamelCase(s string) string {
	parts := strings.Split(s, "-")
	result := parts[0]
	for _, p := range parts[1:] {
		if len(p) > 0 {
			result += capitalize(p)
		}
	}
	return result
}

// capitalize uppercases the first letter.
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// formatVetroValue formats a value for Vetro source.
func formatVetroValue(v string) string {
	if v == "true" || v == "false" {
		return v
	}
	isNum := true
	for _, c := range v {
		if c < '0' || c > '9' {
			isNum = false
			break
		}
	}
	if isNum && len(v) > 0 {
		return v
	}
	return fmt.Sprintf("%q", v)
}

// escapeVetroString escapes special characters for Vetro string literals.
func escapeVetroString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return s
}

// emitMenu emits a <menu> as declarative Vetro syntax.
func (e *VetroEmitter) emitMenu(b *strings.Builder, menu xmlMenu, indent int) error {
	indentStr := strings.Repeat("    ", indent)

	// Menu(id: "...") {
	if menu.ID != "" {
		fmt.Fprintf(b, "%sMenu(id: %q)\n", indentStr, menu.ID)
	} else {
		fmt.Fprintf(b, "%sMenu()\n", indentStr)
	}
	fmt.Fprintf(b, "%s{\n", indentStr)

	// Sections
	for _, section := range menu.Sections {
		fmt.Fprintf(b, "%s    Section\n", indentStr)
		fmt.Fprintf(b, "%s    {\n", indentStr)

		// Items
		for _, item := range section.Items {
			var label, action string
			var translatable []string
			for _, attr := range item.Attributes {
				switch attr.Name {
				case "label":
					label = attr.Value
				case "action":
					action = attr.Value
				}
				if attr.Translatable == "yes" {
					translatable = append(translatable, attr.Name)
				}
			}

			// Item(label: "...", action: "...")
			if action != "" {
				fmt.Fprintf(b, "%s        Item(label: %q, action: %q)\n", indentStr, label, action)
			} else {
				fmt.Fprintf(b, "%s        Item(label: %q)\n", indentStr, label)
			}

			if len(translatable) > 0 {
				args := make([]string, len(translatable))
				for i, p := range translatable {
					args[i] = fmt.Sprintf("%q", p)
				}
				fmt.Fprintf(b, "%s            .translatable(%s)\n", indentStr, strings.Join(args, ", "))
			}
		}

		fmt.Fprintf(b, "%s    }\n", indentStr)
	}

	fmt.Fprintf(b, "%s}\n", indentStr)
	return nil
}

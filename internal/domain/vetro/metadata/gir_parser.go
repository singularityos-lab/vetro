package metadata

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

// GIR structures for XML unmarshaling
type Repository struct {
	XMLName   xml.Name  `xml:"repository"`
	Namespace Namespace `xml:"namespace"`
}

type Namespace struct {
	Name       string  `xml:"name,attr"`
	Version    string  `xml:"version,attr"`
	Classes    []Class `xml:"class"`
	Interfaces []Class `xml:"interface"`
	Aliases    []Alias `xml:"alias"`
	Enums      []Enum  `xml:"enumeration"`
}

type Alias struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

type Enum struct {
	Name    string   `xml:"name,attr"`
	Members []Member `xml:"member"`
}

type Member struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type Property struct {
	Name      string `xml:"name,attr"`
	Type      Type   `xml:"type"`
	Writeable string `xml:"writable,attr"`
}

type Signal struct {
	Name string `xml:"name,attr"`
}

type Type struct {
	Name string `xml:"name,attr"`
}

type Class struct {
	Name       string     `xml:"name,attr"`
	Parent     string     `xml:"parent,attr"`
	Properties []Property `xml:"property"`
	Signals    []Signal   `xml:"http://www.gtk.org/introspection/glib/1.0 signal"`
}

// Metadata is the processed version of GIR data
type Metadata struct {
	Classes map[string]*ClassMetadata `json:"classes"`
}

type ClassMetadata struct {
	Name       string                    `json:"name"`
	Namespace  string                    `json:"namespace"`
	Parent     string                    `json:"parent"`
	Properties map[string]PropertySchema `json:"properties"`
	Signals    []string                  `json:"signals"`
	Resolved   bool                      `json:"-"`
}

// Merge copies all classes from other into m, without overwriting existing ones.
func (m *Metadata) Merge(other *Metadata) {
	for k, v := range other.Classes {
		if _, exists := m.Classes[k]; !exists {
			m.Classes[k] = v
		}
	}
}

type PropertySchema struct {
	Type     string   `json:"type"`
	EnumVals []string `json:"enum_vals,omitempty"`
}

// ResolveInheritance flattens inheritance hierarchies
func (m *Metadata) ResolveInheritance() {
	for _, class := range m.Classes {
		m.resolveClass(class, map[string]bool{})
	}
}

func (m *Metadata) resolveClass(class *ClassMetadata, resolving map[string]bool) {
	if class.Resolved {
		return
	}

	if resolving[class.Name] {
		// GIR data is expected to be acyclic, but guard against malformed metadata.
		class.Resolved = true
		return
	}
	resolving[class.Name] = true
	defer delete(resolving, class.Name)

	if class.Parent == "" {
		class.Resolved = true
		return
	}

	var parentKey string
	if strings.Contains(class.Parent, ".") {
		// Cross-namespace: "Gtk.ApplicationWindow" → "GtkApplicationWindow"
		parts := strings.SplitN(class.Parent, ".", 2)
		parentKey = parts[0] + parts[1]
	} else {
		// Same namespace: resolve using class's own namespace prefix
		parentKey = class.Namespace + class.Parent
	}

	parent, ok := m.Classes[parentKey]
	if !ok {
		class.Resolved = true
		return
	}

	m.resolveClass(parent, resolving)

	// Inherit properties
	for k, v := range parent.Properties {
		if _, exists := class.Properties[k]; !exists {
			class.Properties[k] = v
		}
	}

	// Inherit signals
	signalMap := make(map[string]bool)
	for _, s := range class.Signals {
		signalMap[s] = true
	}
	for _, s := range parent.Signals {
		if !signalMap[s] {
			class.Signals = append(class.Signals, s)
		}
	}

	class.Resolved = true
}

// ParseGIR reads a .gir file and returns the processed metadata
func ParseGIR(path string) (*Metadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var repo Repository
	if err := xml.Unmarshal(data, &repo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GIR: %w", err)
	}

	nsPrefix := repo.Namespace.Name

	meta := &Metadata{
		Classes: make(map[string]*ClassMetadata),
	}

	processClasses(repo.Namespace.Classes, meta, nsPrefix)
	processClasses(repo.Namespace.Interfaces, meta, nsPrefix)

	meta.ResolveInheritance()

	return meta, nil
}

func processClasses(classes []Class, meta *Metadata, nsPrefix string) {
	for _, c := range classes {
		fullClassName := nsPrefix + c.Name
		classMeta := &ClassMetadata{
			Name:       fullClassName,
			Namespace:  nsPrefix,
			Parent:     c.Parent,
			Properties: make(map[string]PropertySchema),
			Signals:    make([]string, 0),
		}

		for _, p := range c.Properties {
			classMeta.Properties[p.Name] = PropertySchema{
				Type: p.Type.Name,
			}
		}

		for _, s := range c.Signals {
			classMeta.Signals = append(classMeta.Signals, s.Name)
		}

		meta.Classes[fullClassName] = classMeta
	}
}

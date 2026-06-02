package vetro

import (
	"os"
	"strings"
)

const gtkClassPrefix = "Gtk"

// isCamelSuffix reports whether fullName ends with name at a CamelCase word
// boundary, i.e. name begins a new capitalised component of fullName (so
// "SingularityWidgetsChip" matches "Chip" but "...somethingchip" would not).
func isCamelSuffix(fullName, name string) bool {
	if !strings.HasSuffix(fullName, name) {
		return false
	}
	idx := len(fullName) - len(name)
	if idx == 0 {
		return true
	}
	prev := fullName[idx-1]
	return (prev >= 'a' && prev <= 'z') || (prev >= '0' && prev <= '9')
}

// ReadFile reads the entire file content.
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteFile writes the content to a file.
func WriteFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func qualifyGtkClassName(name string) string {
	if strings.HasPrefix(name, gtkClassPrefix) {
		return name
	}

	// If MetadataManager is loaded, search all namespaces for a class
	// whose simple name matches (e.g. "TabBar" matches "SingularityTabBar")
	if MetadataManager != nil && MetadataManager.Metadata != nil {
		// Exact full match first (already qualified by caller)
		if _, ok := MetadataManager.Metadata.Classes[name]; ok {
			return name
		}
		// Search by simple name across all namespaces.
		// Priority: (1) Gtk exact, (2) Singularity suffix, (3) other exact, (4) other suffix.
		gtkMatch := ""
		otherMatch := ""
		singularitySuffixMatch := ""
		otherSuffixMatch := ""

		for fullName, class := range MetadataManager.Metadata.Classes {
			simpleName := strings.TrimPrefix(fullName, class.Namespace)
			isGtk := strings.HasPrefix(fullName, gtkClassPrefix)
			isSingularity := strings.HasPrefix(fullName, "Singularity")

			if simpleName == name {
				if isGtk {
					gtkMatch = fullName
				} else if otherMatch == "" {
					otherMatch = fullName
				}
			}
			// Suffix match at a CamelCase boundary (catches sub-namespaced
			// types, e.g. WidgetsToolBar for the short name ToolBar). When
			// several classes share
			// a suffix (e.g. Chip and CalendarEventChip both end in "Chip"),
			// prefer the shortest full name: that is the most exact match, and
			// it is deterministic regardless of Go's random map order.
			if simpleName != name && isCamelSuffix(fullName, name) {
				if isSingularity {
					if singularitySuffixMatch == "" || len(fullName) < len(singularitySuffixMatch) {
						singularitySuffixMatch = fullName
					}
				} else if !isGtk {
					if otherSuffixMatch == "" || len(fullName) < len(otherSuffixMatch) {
						otherSuffixMatch = fullName
					}
				}
			}
		}

		// Return in priority order
		if gtkMatch != "" {
			return gtkMatch
		}
		if singularitySuffixMatch != "" {
			return singularitySuffixMatch
		}
		if otherMatch != "" {
			return otherMatch
		}
		if otherSuffixMatch != "" {
			return otherSuffixMatch
		}
	}

	return gtkClassPrefix + name
}

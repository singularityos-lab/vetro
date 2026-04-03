package vetro

// ComponentNode represents a node in the AST.
type ComponentNode struct {
	Type         string                // e.g., "Window", "VBox"
	Properties   map[string]any        // e.g., map[string]any{"title": "La mia App", "default-width": 800}
	Signals      map[string]string     // e.g., map[string]string{"clicked": "my_click_handler"}
	SignalAttrs  map[string]SignalAttr // signal-specific attributes (swapped, after)
	ChildType    string                // e.g., "titlebar" for special child types
	ChildName    string                // e.g., "page1" for Stack pages
	Translatable []string              // property names marked as translatable
	Children     []*ComponentNode      // child components
}

// SignalAttr holds extra attributes for a signal.
type SignalAttr struct {
	Swapped bool // swapped="no" or "yes"
	After   bool // after="yes" (emit after default handler)
}

// StyleBlock represents an inline CSS style block.
type StyleBlock struct {
	CSS      string // CSS content
	Priority string // "application", "user", "fallback"
}

// Requires represents a GTK version requirement.
type Requires struct {
	Lib     string // e.g., "gtk"
	Version string // e.g., "4.0"
}

// Program represents the complete parsed program (root AST node).
type Program struct {
	Roots    []*ComponentNode // top-level UI components (orphan objects + main root)
	Styles   []StyleBlock     // CSS style blocks
	Menus    []*ComponentNode // top-level menu definitions
	Requires *Requires        // GTK version requirement
}

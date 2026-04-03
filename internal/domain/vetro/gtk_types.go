package vetro

import (
	"vetro/internal/domain/vetro/metadata"
)

// MetadataManager is the global instance for dynamic GTK metadata.
var MetadataManager *metadata.Manager

// PropertyType represents the expected type of a GTK property value.
type PropertyType int

const (
	PropTypeString PropertyType = iota
	PropTypeInt
	PropTypeBool
	PropTypeEnum
)

// PropertySchema defines the expected type for a GTK property.
type PropertySchema struct {
	Type     PropertyType
	EnumVals []string
}

// gtkPropertyTypes maps GTK property names to their expected types.
var gtkPropertyTypes = map[string]PropertySchema{
	// Window properties
	"title":               {Type: PropTypeString},
	"default-width":       {Type: PropTypeInt},
	"default-height":      {Type: PropTypeInt},
	"resizable":           {Type: PropTypeBool},
	"modal":               {Type: PropTypeBool},
	"destroy-with-parent": {Type: PropTypeBool},
	"icon-name":           {Type: PropTypeString},

	// Widget properties
	"halign":      {Type: PropTypeEnum, EnumVals: []string{"fill", "start", "end", "center", "baseline"}},
	"valign":      {Type: PropTypeEnum, EnumVals: []string{"fill", "start", "end", "center", "baseline"}},
	"hexpand":     {Type: PropTypeBool},
	"vexpand":     {Type: PropTypeBool},
	"sensitive":   {Type: PropTypeBool},
	"visible":     {Type: PropTypeBool},
	"opacity":     {Type: PropTypeInt},
	"css-classes": {Type: PropTypeString},

	// Box properties
	"orientation":       {Type: PropTypeEnum, EnumVals: []string{"horizontal", "vertical"}},
	"spacing":           {Type: PropTypeInt},
	"homogeneous":       {Type: PropTypeBool},
	"baseline-position": {Type: PropTypeEnum, EnumVals: []string{"top", "center", "bottom"}},

	// Margin properties
	"margin-top":    {Type: PropTypeInt},
	"margin-bottom": {Type: PropTypeInt},
	"margin-start":  {Type: PropTypeInt},
	"margin-end":    {Type: PropTypeInt},

	// Label properties
	"label":           {Type: PropTypeString},
	"use-markup":      {Type: PropTypeBool},
	"use-underline":   {Type: PropTypeBool},
	"wrap":            {Type: PropTypeBool},
	"wrap-mode":       {Type: PropTypeEnum, EnumVals: []string{"none", "char", "word", "word-char"}},
	"justify":         {Type: PropTypeEnum, EnumVals: []string{"left", "right", "center", "fill"}},
	"selectable":      {Type: PropTypeBool},
	"lines":           {Type: PropTypeInt},
	"max-width-chars": {Type: PropTypeInt},

	// Button properties
	"child": {Type: PropTypeString},

	// Entry properties
	"placeholder-text":  {Type: PropTypeString},
	"max-length":        {Type: PropTypeInt},
	"visibility":        {Type: PropTypeBool},
	"editable":          {Type: PropTypeBool},
	"activates-default": {Type: PropTypeBool},
	"input-purpose":     {Type: PropTypeEnum, EnumVals: []string{"free-form", "alpha", "digits", "number", "phone", "url", "email", "name", "password", "pin"}},
	"input-hints":       {Type: PropTypeString},

	// Switch properties
	"active": {Type: PropTypeBool},
	"state":  {Type: PropTypeBool},

	// Scale properties
	"value":          {Type: PropTypeInt},
	"lower":          {Type: PropTypeInt},
	"upper":          {Type: PropTypeInt},
	"step-increment": {Type: PropTypeInt},
	"page-increment": {Type: PropTypeInt},
	"digits":         {Type: PropTypeInt},

	// HeaderBar properties
	"show-title-buttons": {Type: PropTypeBool},
	"title-widget":       {Type: PropTypeString},
	"decoration-layout":  {Type: PropTypeString},

	// ScrolledWindow properties
	"policy":             {Type: PropTypeEnum, EnumVals: []string{"always", "automatic", "never"}},
	"kinetic-scrolling":  {Type: PropTypeBool},
	"overlay-scrolling":  {Type: PropTypeBool},
	"min-content-width":  {Type: PropTypeInt},
	"min-content-height": {Type: PropTypeInt},

	// Grid properties
	"row-spacing":        {Type: PropTypeInt},
	"column-spacing":     {Type: PropTypeInt},
	"row-homogeneous":    {Type: PropTypeBool},
	"column-homogeneous": {Type: PropTypeBool},

	// Stack properties
	"transition-type":     {Type: PropTypeEnum, EnumVals: []string{"none", "crossfade", "slide-right", "slide-left", "slide-up", "slide-down"}},
	"transition-duration": {Type: PropTypeInt},
	"hhomogeneous":        {Type: PropTypeBool},
	"vhomogeneous":        {Type: PropTypeBool},

	// ListBox properties
	"selection-mode":           {Type: PropTypeEnum, EnumVals: []string{"none", "single", "browse", "multiple"}},
	"activate-on-single-click": {Type: PropTypeBool},

	// Image properties
	"file":       {Type: PropTypeString},
	"pixel-size": {Type: PropTypeInt},
	"icon-size":  {Type: PropTypeEnum, EnumVals: []string{"inherit", "normal", "large"}},

	// ProgressBar properties
	"fraction":   {Type: PropTypeInt},
	"pulse-step": {Type: PropTypeInt},
	"text":       {Type: PropTypeString},
	"show-text":  {Type: PropTypeBool},
	"inverted":   {Type: PropTypeBool},

	// Spinner properties
	"spinning": {Type: PropTypeBool},

	// Calendar properties
	"show-heading":      {Type: PropTypeBool},
	"show-day-names":    {Type: PropTypeBool},
	"show-week-numbers": {Type: PropTypeBool},

	// Menu properties (GMenu)
	"action":          {Type: PropTypeString},
	"section-name":    {Type: PropTypeString},
	"action-name":     {Type: PropTypeString},
	"action-target":   {Type: PropTypeString},
	"detailed-action": {Type: PropTypeString},
	"menu-model":      {Type: PropTypeString},

	// Accessibility properties
	"accessible-label":       {Type: PropTypeString},
	"accessible-description": {Type: PropTypeString},
	"accessible-role":        {Type: PropTypeString},
	"accessible-labeled-by":  {Type: PropTypeString},

	// Grid child properties
	"row":         {Type: PropTypeInt},
	"column":      {Type: PropTypeInt},
	"row-span":    {Type: PropTypeInt},
	"column-span": {Type: PropTypeInt},
	"left-attach": {Type: PropTypeInt},
	"top-attach":  {Type: PropTypeInt},
	"width":       {Type: PropTypeInt},
	"height":      {Type: PropTypeInt},

	// Popover properties
	"autohide":        {Type: PropTypeBool},
	"cascade-popdown": {Type: PropTypeBool},
	"has-arrow":       {Type: PropTypeBool},
	"pointing-to":     {Type: PropTypeString},
	"relative-to":     {Type: PropTypeString},
}

// LookupPropertySchema returns the schema for a GTK property, or nil if unknown.
func LookupPropertySchema(gtkPropName string) *PropertySchema {
	// Try dynamic metadata first if available
	if MetadataManager != nil && MetadataManager.Metadata != nil {
		for _, class := range MetadataManager.Metadata.Classes {
			if prop, ok := class.Properties[gtkPropName]; ok {
				return &PropertySchema{
					Type: mapGirTypeToPropType(prop.Type),
				}
			}
		}
	}

	if schema, ok := gtkPropertyTypes[gtkPropName]; ok {
		return &schema
	}
	return nil
}

func mapGirTypeToPropType(girType string) PropertyType {
	switch girType {
	case "gboolean":
		return PropTypeBool
	case "gint", "gdouble", "gfloat", "guint", "glong", "gulong":
		return PropTypeInt
	default:
		return PropTypeString
	}
}

// gtkValidSignals maps GTK widget types to their valid signal names.
var gtkValidSignals = map[string][]string{
	"GtkButton":      {"clicked", "pressed", "released", "activate"},
	"GtkEntry":       {"changed", "activate", "insert-text", "delete-text"},
	"GtkWindow":      {"close-request", "destroy", "show", "hide", "focus-in-event"},
	"GtkLabel":       {"activate-link", "copy-clipboard"},
	"GtkSwitch":      {"state-set", "notify::active"},
	"GtkScale":       {"value-changed", "change-value"},
	"GtkCheckButton": {"toggled", "activate"},
	"GtkSpinButton":  {"value-changed", "input", "output", "wrapped"},
	"GtkComboBox":    {"changed"},
	"GtkListBox":     {"row-selected", "row-activated"},
	"GtkStack":       {"notify::visible-child"},
}

// LookupValidSignals returns valid signal names for a GTK widget type.
func LookupValidSignals(gtkClassName string) []string {
	if MetadataManager != nil && MetadataManager.Metadata != nil {
		if class, ok := MetadataManager.Metadata.Classes[gtkClassName]; ok {
			return class.Signals
		}
	}

	if signals, ok := gtkValidSignals[gtkClassName]; ok {
		return signals
	}
	return nil
}

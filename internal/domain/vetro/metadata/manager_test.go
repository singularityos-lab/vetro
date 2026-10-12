package metadata

import (
	"os"
	"path/filepath"
	"testing"
)

func writeGIR(t *testing.T, dir, name, source string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestGIRPathPriority(t *testing.T) {
	local, system := t.TempDir(), t.TempDir()
	writeGIR(t, local, "Gtk-4.0.gir", `<repository><namespace name="Gtk" version="4.0"><class name="Button"><property name="sit-property" writable="1"><type name="gboolean"/></property></class></namespace></repository>`)
	writeGIR(t, system, "Gtk-4.0.gir", `<repository><namespace name="Gtk" version="4.0"><class name="Button"><property name="system-property" writable="1"><type name="gboolean"/></property></class></namespace></repository>`)
	m := &Manager{CacheDir: t.TempDir()}
	if err := m.LoadGIRPaths([]string{local, system}, false); err != nil {
		t.Fatal(err)
	}
	properties := m.Metadata.Classes["GtkButton"].Properties
	if _, ok := properties["sit-property"]; !ok {
		t.Fatal("local toolkit property missing")
	}
	if _, ok := properties["system-property"]; ok {
		t.Fatal("system metadata replaced local toolkit metadata")
	}
	if err := m.LoadGIRPaths([]string{local, system}, false); err != nil {
		t.Fatal(err)
	}
	if _, ok := m.Metadata.Classes["GtkButton"].Properties["sit-property"]; !ok {
		t.Fatal("cached toolkit metadata lost priority")
	}
}

func TestGIRPathErrors(t *testing.T) {
	m := &Manager{CacheDir: t.TempDir()}
	empty := t.TempDir()
	if err := m.LoadGIRPaths([]string{empty}, false); err == nil {
		t.Fatal("empty explicit GIR directory accepted")
	}
	if err := m.LoadGIRPaths([]string{filepath.Join(empty, "missing")}, false); err == nil {
		t.Fatal("missing explicit GIR directory accepted")
	}
	writeGIR(t, empty, "Gtk-4.0.gir", `<repository`)
	if err := m.LoadGIRPaths([]string{empty}, false); err == nil {
		t.Fatal("malformed explicit GIR accepted")
	}
	writeGIR(t, empty, "Gtk-4.0.gir", `<repository/>`)
	if err := m.LoadGIRPaths([]string{empty}, false); err == nil {
		t.Fatal("GIR without a namespace accepted")
	}
}

func TestLateCrossNamespaceParent(t *testing.T) {
	dir := t.TempDir()
	writeGIR(t, dir, "App-1.0.gir", `<repository><namespace name="App"><class name="Window" parent="Gtk.Window"/></namespace></repository>`)
	writeGIR(t, dir, "Gtk-4.0.gir", `<repository><namespace name="Gtk"><class name="Window"><property name="title" writable="1"><type name="utf8"/></property></class></namespace></repository>`)
	m := &Manager{CacheDir: t.TempDir()}
	if err := m.LoadGIRPaths([]string{dir}, false); err != nil {
		t.Fatal(err)
	}
	m.ResolveInheritance()
	if _, ok := m.Metadata.Classes["AppWindow"].Properties["title"]; !ok {
		t.Fatal("parent loaded later was not inherited")
	}
}

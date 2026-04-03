package vetro

import (
	"strings"
	"testing"
)

func TestVetroEmitter_TranslatableMenu(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<interface>
  <menu id="app-menu">
    <section>
      <item>
        <attribute name="label" translatable="yes">Open</attribute>
        <attribute name="action">app.open</attribute>
      </item>
    </section>
  </menu>
</interface>`

	emitter := NewVetroEmitter()
	output, err := emitter.EmitXMLToVetro(xml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `.translatable("label")`
	if !strings.Contains(output, expected) {
		t.Errorf("expected translatable modifier in output, got:\n%s", output)
	}
}

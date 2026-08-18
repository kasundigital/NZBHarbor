package nzb

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	x := `<?xml version="1.0"?><nzb><file subject="post &quot;example.bin&quot;" date="1"><groups><group>alt.test</group></groups><segments><segment bytes="3" number="2">b@test</segment><segment bytes="3" number="1">a@test</segment></segments></file></nzb>`
	doc, err := Parse(strings.NewReader(x))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Files) != 1 || doc.Files[0].Filename != "example.bin" {
		t.Fatalf("unexpected file: %+v", doc.Files)
	}
	if doc.Files[0].Segments[0].Number != 1 {
		t.Fatalf("segments not sorted")
	}
}

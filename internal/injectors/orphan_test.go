package injectors

import (
	"testing"

	"github.com/novusedge/hirejack/internal/pdf"
)

func TestOrphanInjector_InjectPayload(t *testing.T) {
	p, _ := pdf.Parse(pdf.CreateTestPDF())
	before := p.ObjectCount()

	injector := NewOrphanInjector()
	err := injector.InjectPayload(p, []byte("hidden payload"))
	if err != nil {
		t.Fatalf("InjectPayload: %v", err)
	}

	after := p.ObjectCount()
	if after <= before {
		t.Error("object count should increase")
	}
}

func TestOrphanInjector_InjectMultiple(t *testing.T) {
	p, _ := pdf.Parse(pdf.CreateTestPDF())
	before := p.ObjectCount()

	injector := NewOrphanInjector()
	payloads := [][]byte{
		[]byte("payload 1"),
		[]byte("payload 2"),
		[]byte("payload 3"),
	}

	err := injector.InjectMultiple(p, payloads)
	if err != nil {
		t.Fatalf("InjectMultiple: %v", err)
	}

	after := p.ObjectCount()
	if after-before < 3 {
		t.Errorf("should add 3 objects, added %d", after-before)
	}
}

func TestOrphanInjector_Inject(t *testing.T) {
	p, _ := pdf.Parse(pdf.CreateTestPDF())
	before := p.ObjectCount()

	injector := NewOrphanInjector()
	err := injector.Inject(p, OrphanConfig{
		Content: []byte("single payload"),
		Multiple: [][]byte{
			[]byte("multi 1"),
			[]byte("multi 2"),
		},
	})
	if err != nil {
		t.Fatalf("Inject: %v", err)
	}

	after := p.ObjectCount()
	if after-before < 3 {
		t.Errorf("should add 3+ objects, added %d", after-before)
	}
}

func TestOrphanInjector_WriteSucceeds(t *testing.T) {
	p, _ := pdf.Parse(pdf.CreateTestPDF())

	injector := NewOrphanInjector()
	injector.InjectPayload(p, []byte("orphan content"))

	// Write should succeed even with orphan objects
	out, err := p.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	// Should still be valid PDF
	p2, err := pdf.Parse(out)
	if err != nil {
		t.Fatalf("reparse: %v", err)
	}
	if !p2.IsValid() {
		t.Error("PDF should be valid after orphan injection")
	}
}

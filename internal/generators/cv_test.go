package generators

import (
	"strings"
	"testing"

	"github.com/novusedge/hirejack/internal/pdf"
)

func TestCVGenerate(t *testing.T) {
	gen := NewCVGenerator(nil)
	pgen := NewPersonaGenerator()
	p := pgen.Generate()

	doc, err := gen.Generate(p)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if doc == nil {
		t.Fatal("doc should not be nil")
	}

	info, _ := doc.GetInfoDict()
	if info.Author != p.FullName() {
		t.Errorf("Author = %q, want %q", info.Author, p.FullName())
	}
}

func TestGenerateBytes(t *testing.T) {
	gen := NewCVGenerator(nil)
	pgen := NewPersonaGenerator()
	p := pgen.Generate()

	data, err := gen.GenerateBytes(p)
	if err != nil {
		t.Fatalf("GenerateBytes: %v", err)
	}

	if len(data) == 0 {
		t.Error("should return PDF bytes")
	}

	// Should be reparseable
	_, err = pdf.Parse(data)
	if err != nil {
		t.Errorf("reparse: %v", err)
	}
}

func TestGenerateWithContent(t *testing.T) {
	gen := NewCVGenerator(nil)
	pgen := NewPersonaGenerator()
	p := pgen.Generate()

	doc, err := gen.GenerateWithContent(p)
	if err != nil {
		t.Fatalf("GenerateWithContent: %v", err)
	}

	if doc == nil {
		t.Fatal("doc should not be nil")
	}
}

func TestMetadataSet(t *testing.T) {
	gen := NewCVGenerator(nil)
	p := &Persona{
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
		Phone:     "(555) 123-4567",
		City:      "Test City",
		Title:     "Software Engineer",
		Summary:   "Test summary",
		Skills:    []string{"Go", "Python"},
	}

	doc, _ := gen.Generate(p)
	info, _ := doc.GetInfoDict()

	if info.Author != "Test User" {
		t.Errorf("Author = %q", info.Author)
	}
	if !strings.Contains(info.Title, "Resume") {
		t.Errorf("Title = %q", info.Title)
	}
	if info.Subject != "Software Engineer" {
		t.Errorf("Subject = %q", info.Subject)
	}
	if !strings.Contains(info.Keywords, "Go") {
		t.Errorf("Keywords = %q", info.Keywords)
	}
	if info.Custom["Email"] != "test@example.com" {
		t.Errorf("Custom[Email] = %q", info.Custom["Email"])
	}
}

func TestXMPSet(t *testing.T) {
	gen := NewCVGenerator(nil)
	p := &Persona{
		FirstName: "Jane",
		LastName:  "Doe",
		Title:     "Developer",
		Summary:   "Experienced developer",
	}

	doc, _ := gen.Generate(p)
	xmp, _ := doc.GetXMP()

	if xmp == nil {
		t.Fatal("XMP should be set")
	}
}

func TestCustomBasePDF(t *testing.T) {
	customBase := pdf.CreateTestPDF()
	gen := NewCVGenerator(customBase)
	pgen := NewPersonaGenerator()
	p := pgen.Generate()

	doc, err := gen.Generate(p)
	if err != nil {
		t.Fatalf("Generate with custom base: %v", err)
	}

	if doc == nil {
		t.Error("should work with custom base PDF")
	}
}

func TestBuildCVContent(t *testing.T) {
	p := &Persona{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     "(555) 555-5555",
		City:      "NYC",
		Title:     "Engineer",
		Summary:   "Good engineer",
		Experience: []Job{
			{Company: "Corp", Title: "Dev", Duration: "2020-2022", Desc: "Did stuff"},
		},
		Education: []School{
			{Name: "MIT", Degree: "BS CS", Year: "2019"},
		},
		Skills: []string{"Go", "Rust"},
	}

	content := buildCVContent(p)

	if !strings.Contains(string(content), "John Doe") {
		t.Error("should contain name")
	}
	if !strings.Contains(string(content), "john@example.com") {
		t.Error("should contain email")
	}
	if !strings.Contains(string(content), "Corp") {
		t.Error("should contain company")
	}
	if !strings.Contains(string(content), "MIT") {
		t.Error("should contain school")
	}
	if !strings.Contains(string(content), "Go, Rust") {
		t.Error("should contain skills")
	}
}

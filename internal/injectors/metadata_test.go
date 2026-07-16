package injectors

import (
	"testing"

	"github.com/novusedge/hirejack/internal/pdf"
)

func TestMetadataInjector_Inject(t *testing.T) {
	p, err := pdf.Parse(pdf.CreateTestPDF())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	injector := NewMetadataInjector()

	err = injector.Inject(p, MetadataConfig{
		Author: "'; DROP TABLE users; --",
		Title:  "Test Resume",
	})
	if err != nil {
		t.Fatalf("Inject: %v", err)
	}

	info, err := p.GetInfoDict()
	if err != nil {
		t.Fatalf("GetInfoDict: %v", err)
	}
	if info.Author != "'; DROP TABLE users; --" {
		t.Errorf("Author = %q, want SQL payload", info.Author)
	}
	if info.Title != "Test Resume" {
		t.Errorf("Title = %q, want %q", info.Title, "Test Resume")
	}
}

func TestMetadataInjector_InjectMap(t *testing.T) {
	p, _ := pdf.Parse(pdf.CreateTestPDF())
	injector := NewMetadataInjector()

	fields := map[string]string{
		"Author":      "payload1",
		"Title":       "payload2",
		"CustomField": "payload3",
	}

	err := injector.InjectMap(p, fields)
	if err != nil {
		t.Fatalf("InjectMap: %v", err)
	}

	info, _ := p.GetInfoDict()
	if info.Author != "payload1" {
		t.Errorf("Author = %q, want %q", info.Author, "payload1")
	}
	if info.Custom["CustomField"] != "payload3" {
		t.Errorf("Custom[CustomField] = %q, want %q", info.Custom["CustomField"], "payload3")
	}
}

func TestMetadataInjector_InjectField(t *testing.T) {
	p, _ := pdf.Parse(pdf.CreateTestPDF())
	injector := NewMetadataInjector()

	err := injector.InjectField(p, "Subject", "test subject")
	if err != nil {
		t.Fatalf("InjectField: %v", err)
	}

	info, _ := p.GetInfoDict()
	if info.Subject != "test subject" {
		t.Errorf("Subject = %q, want %q", info.Subject, "test subject")
	}
}

func TestMetadataInjector_PreservesExisting(t *testing.T) {
	p, _ := pdf.Parse(pdf.CreateTestPDF())
	injector := NewMetadataInjector()

	// Set initial
	injector.Inject(p, MetadataConfig{Author: "first", Title: "original"})

	// Update only author
	injector.Inject(p, MetadataConfig{Author: "second"})

	info, _ := p.GetInfoDict()
	if info.Author != "second" {
		t.Errorf("Author = %q, want %q", info.Author, "second")
	}
	if info.Title != "original" {
		t.Errorf("Title should be preserved: %q", info.Title)
	}
}

func TestMetadataInjector_AllFields(t *testing.T) {
	p, _ := pdf.Parse(pdf.CreateTestPDF())
	injector := NewMetadataInjector()

	err := injector.Inject(p, MetadataConfig{
		Author:   "a",
		Title:    "t",
		Subject:  "s",
		Keywords: "k",
		Creator:  "c",
		Producer: "p",
	})
	if err != nil {
		t.Fatalf("Inject: %v", err)
	}

	info, _ := p.GetInfoDict()
	if info.Author != "a" || info.Title != "t" || info.Subject != "s" ||
		info.Keywords != "k" || info.Creator != "c" || info.Producer != "p" {
		t.Error("not all fields set correctly")
	}
}

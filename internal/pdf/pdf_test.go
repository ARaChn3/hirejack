package pdf

import (
	"testing"
)

func TestParse(t *testing.T) {
	data := CreateTestPDF()
	p, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if p == nil {
		t.Fatal("Parse returned nil")
	}
}

func TestParseEmpty(t *testing.T) {
	_, err := Parse(nil)
	if err == nil {
		t.Error("Parse(nil) should error")
	}
	_, err = Parse([]byte{})
	if err == nil {
		t.Error("Parse([]) should error")
	}
}

func TestBytes(t *testing.T) {
	p, err := Parse(CreateTestPDF())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	out, err := p.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	if len(out) == 0 {
		t.Error("Bytes returned empty")
	}
	// Should be reparseable
	_, err = Parse(out)
	if err != nil {
		t.Errorf("reparse: %v", err)
	}
}

func TestIsValid(t *testing.T) {
	p, _ := Parse(CreateTestPDF())
	if !p.IsValid() {
		t.Error("test PDF should be valid")
	}
}

func TestInfoDict(t *testing.T) {
	p, _ := Parse(CreateTestPDF())

	// Initially nil or empty
	info, err := p.GetInfoDict()
	if err != nil {
		t.Fatalf("GetInfoDict: %v", err)
	}

	// Set info
	err = p.SetInfoDict(&InfoDict{
		Author:  "test author",
		Title:   "test title",
		Subject: "test subject",
	})
	if err != nil {
		t.Fatalf("SetInfoDict: %v", err)
	}

	// Read back
	info, err = p.GetInfoDict()
	if err != nil {
		t.Fatalf("GetInfoDict: %v", err)
	}
	if info == nil {
		t.Fatal("info should not be nil")
	}
	if info.Author != "test author" {
		t.Errorf("Author = %q, want %q", info.Author, "test author")
	}
	if info.Title != "test title" {
		t.Errorf("Title = %q, want %q", info.Title, "test title")
	}
}

func TestInfoDictPreserved(t *testing.T) {
	p, _ := Parse(CreateTestPDF())
	p.SetInfoDict(&InfoDict{Author: "preserved"})

	out, _ := p.Bytes()
	p2, _ := Parse(out)
	info, _ := p2.GetInfoDict()

	if info == nil || info.Author != "preserved" {
		t.Error("info dict not preserved through write/parse cycle")
	}
}

func TestXMP(t *testing.T) {
	p, _ := Parse(CreateTestPDF())

	// Initially nil
	xmp, err := p.GetXMP()
	if err != nil {
		t.Fatalf("GetXMP: %v", err)
	}
	if xmp != nil {
		t.Log("PDF has existing XMP")
	}

	// Set XMP
	newXMP := NewXMPPacket()
	err = p.SetXMP(newXMP)
	if err != nil {
		t.Fatalf("SetXMP: %v", err)
	}

	// Read back
	xmp, err = p.GetXMP()
	if err != nil {
		t.Fatalf("GetXMP after set: %v", err)
	}
	if xmp == nil {
		t.Fatal("XMP should exist after SetXMP")
	}
	if len(xmp.Raw) == 0 {
		t.Error("XMP.Raw should not be empty")
	}
}

func TestOrphanStream(t *testing.T) {
	p, _ := Parse(CreateTestPDF())
	before := p.ObjectCount()

	err := p.CreateOrphanStream([]byte("hidden content"))
	if err != nil {
		t.Fatalf("CreateOrphanStream: %v", err)
	}

	after := p.ObjectCount()
	if after <= before {
		t.Error("object count should increase")
	}
}

func TestAppendIncremental(t *testing.T) {
	p, _ := Parse(CreateTestPDF())

	out, err := p.AppendIncremental(map[string]string{
		"Author": "incremental author",
	})
	if err != nil {
		t.Fatalf("AppendIncremental: %v", err)
	}

	p2, _ := Parse(out)
	info, _ := p2.GetInfoDict()
	if info == nil || info.Author != "incremental author" {
		t.Error("incremental update not applied")
	}
}

func TestObjectCount(t *testing.T) {
	p, _ := Parse(CreateTestPDF())
	if p.ObjectCount() < 3 {
		t.Errorf("ObjectCount = %d, want >= 3", p.ObjectCount())
	}
}

func TestRaw(t *testing.T) {
	data := CreateTestPDF()
	p, _ := Parse(data)
	if len(p.Raw()) != len(data) {
		t.Error("Raw should return original data")
	}
}

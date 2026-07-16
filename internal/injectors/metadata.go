// Package injectors provides payload injection into various PDF locations.
package injectors

import (
	"strings"

	"github.com/novusedge/hirejack/internal/pdf"
)

// MetadataInjector injects payloads into Document Info Dict.
type MetadataInjector struct{}

// MetadataConfig specifies what to inject.
type MetadataConfig struct {
	Author   string `json:"author,omitempty"`
	Title    string `json:"title,omitempty"`
	Subject  string `json:"subject,omitempty"`
	Keywords string `json:"keywords,omitempty"`
	Creator  string `json:"creator,omitempty"`
	Producer string `json:"producer,omitempty"`
}

// NewMetadataInjector creates a new metadata injector.
func NewMetadataInjector() *MetadataInjector {
	return &MetadataInjector{}
}

// Inject applies metadata payloads to a PDF.
func (i *MetadataInjector) Inject(p *pdf.PDF, config MetadataConfig) error {
	info, err := p.GetInfoDict()
	if err != nil {
		return err
	}
	if info == nil {
		info = &pdf.InfoDict{Custom: make(map[string]string)}
	}

	if config.Author != "" {
		info.Author = config.Author
	}
	if config.Title != "" {
		info.Title = config.Title
	}
	if config.Subject != "" {
		info.Subject = config.Subject
	}
	if config.Keywords != "" {
		info.Keywords = config.Keywords
	}
	if config.Creator != "" {
		info.Creator = config.Creator
	}
	if config.Producer != "" {
		info.Producer = config.Producer
	}

	return p.SetInfoDict(info)
}

// InjectField injects a single field.
func (i *MetadataInjector) InjectField(p *pdf.PDF, field, value string) error {
	return i.InjectMap(p, map[string]string{field: value})
}

// InjectMap injects multiple fields from a map.
func (i *MetadataInjector) InjectMap(p *pdf.PDF, fields map[string]string) error {
	info, err := p.GetInfoDict()
	if err != nil {
		return err
	}
	if info == nil {
		info = &pdf.InfoDict{Custom: make(map[string]string)}
	}

	for field, value := range fields {
		switch strings.ToLower(field) {
		case "author":
			info.Author = value
		case "title":
			info.Title = value
		case "subject":
			info.Subject = value
		case "keywords":
			info.Keywords = value
		case "creator":
			info.Creator = value
		case "producer":
			info.Producer = value
		default:
			if info.Custom == nil {
				info.Custom = make(map[string]string)
			}
			info.Custom[field] = value
		}
	}

	return p.SetInfoDict(info)
}

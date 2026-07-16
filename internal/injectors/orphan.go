package injectors

import (
	"github.com/novusedge/hirejack/internal/pdf"
)

// OrphanInjector creates unreferenced objects in the PDF.
// These objects exist in the xref table but aren't reachable from
// the document tree — invisible to viewers but extractable by parsers.
type OrphanInjector struct{}

// OrphanConfig specifies orphan object injection.
type OrphanConfig struct {
	Content  []byte `json:"content,omitempty"`  // Raw content for stream
	Multiple [][]byte `json:"multiple,omitempty"` // Multiple orphan objects
}

// NewOrphanInjector creates a new orphan injector.
func NewOrphanInjector() *OrphanInjector {
	return &OrphanInjector{}
}

// Inject creates orphan stream objects in the PDF.
func (i *OrphanInjector) Inject(p *pdf.PDF, config OrphanConfig) error {
	if len(config.Content) > 0 {
		if err := p.CreateOrphanStream(config.Content); err != nil {
			return err
		}
	}

	for _, content := range config.Multiple {
		if err := p.CreateOrphanStream(content); err != nil {
			return err
		}
	}

	return nil
}

// InjectPayload creates a single orphan stream with the given payload.
func (i *OrphanInjector) InjectPayload(p *pdf.PDF, payload []byte) error {
	return p.CreateOrphanStream(payload)
}

// InjectMultiple creates multiple orphan streams.
func (i *OrphanInjector) InjectMultiple(p *pdf.PDF, payloads [][]byte) error {
	for _, payload := range payloads {
		if err := p.CreateOrphanStream(payload); err != nil {
			return err
		}
	}
	return nil
}

// Package pdf wraps pdfcpu to provide the high-level PDF manipulation
// primitives shared by every payload injector in hirejack.
package pdf

import (
	"bytes"
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// PDF wraps a pdfcpu context for manipulation.
type PDF struct {
	ctx *model.Context
	raw []byte // Original bytes, kept around for incremental updates.
}

// Parse reads PDF bytes and returns a manipulable PDF.
//
// Parsing is intentionally permissive (relaxed validation): the whole point
// of this package is to build and inspect PDFs that are, or will become,
// structurally unusual. Use IsValid to check spec compliance explicitly.
func Parse(data []byte) (*PDF, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("pdf: parse: empty input")
	}

	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	ctx, err := api.ReadContext(bytes.NewReader(data), conf)
	if err != nil {
		return nil, fmt.Errorf("pdf: parse: %w", err)
	}

	raw := make([]byte, len(data))
	copy(raw, data)

	return &PDF{ctx: ctx, raw: raw}, nil
}

// Bytes returns the modified PDF as bytes, fully rewritten from the current
// in-memory object graph.
func (p *PDF) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := api.Write(p.ctx, &buf, p.ctx.Configuration); err != nil {
		return nil, fmt.Errorf("pdf: write: %w", err)
	}
	return buf.Bytes(), nil
}

// IsValid checks if the PDF is structurally valid per the PDF spec.
func (p *PDF) IsValid() bool {
	return api.ValidateContext(p.ctx) == nil
}

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
	// Disable optimization to preserve orphan streams
	conf.Optimize = false
	conf.OptimizeBeforeWriting = false
	conf.OptimizeResourceDicts = false

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

// BytesWithOrphans returns the PDF with orphan streams appended.
// Orphan streams are inserted before %%EOF so they exist in the file
// but aren't referenced from the document tree.
func (p *PDF) BytesWithOrphans() ([]byte, error) {
	pdfBytes, err := p.Bytes()
	if err != nil {
		return nil, err
	}

	if len(orphanBuf) == 0 {
		return pdfBytes, nil
	}

	// Find %%EOF and insert orphans before it
	eofIdx := bytes.LastIndex(pdfBytes, []byte("%%EOF"))
	if eofIdx == -1 {
		eofIdx = len(pdfBytes)
	}

	var result bytes.Buffer
	result.Write(pdfBytes[:eofIdx])

	// Get next object number from xref
	objNum := p.ObjectCount() + 1

	// Append each orphan as a raw stream object
	for _, o := range orphanBuf {
		result.WriteString(fmt.Sprintf("\n%d 0 obj\n", objNum))
		result.WriteString(fmt.Sprintf("<</Length %d>>\n", len(o.content)))
		result.WriteString("stream\n")
		result.Write(o.content)
		result.WriteString("\nendstream\nendobj\n")
		objNum++
	}

	result.WriteString("\n%%EOF\n")

	// Clear buffer for next use
	ResetOrphanBuffer()

	return result.Bytes(), nil
}

// IsValid checks if the PDF is structurally valid per the PDF spec.
func (p *PDF) IsValid() bool {
	return api.ValidateContext(p.ctx) == nil
}

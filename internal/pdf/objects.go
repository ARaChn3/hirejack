package pdf

import (
	"bytes"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// orphanStream holds raw bytes to append after PDF write
type orphanStream struct {
	content []byte
}

// orphanStreams collects streams to append
var orphanBuf []orphanStream

// CreateOrphanStream creates an unreferenced stream object in the PDF.
// Streams are collected and appended raw to preserve them.
func (p *PDF) CreateOrphanStream(content []byte) error {
	orphanBuf = append(orphanBuf, orphanStream{content: content})
	return nil
}

// CreateOrphanStreamRaw is the same as CreateOrphanStream (both are raw now).
func (p *PDF) CreateOrphanStreamRaw(content []byte) error {
	return p.CreateOrphanStream(content)
}

// ResetOrphanBuffer clears the orphan stream buffer.
func ResetOrphanBuffer() {
	orphanBuf = nil
}

// GetOrphanCount returns the number of pending orphan streams.
func GetOrphanCount() int {
	return len(orphanBuf)
}

// AppendIncremental appends an incremental update section to the PDF.
// This creates a new xref/trailer at the end without rewriting the file,
// which can hide modifications from naive parsers that only read the
// first xref table.
func (p *PDF) AppendIncremental(updates map[string]string) ([]byte, error) {
	// ponytail: minimal incremental — just append info dict updates
	// upgrade: proper xref table generation if needed

	if len(updates) == 0 {
		return p.raw, nil
	}

	info, err := p.GetInfoDict()
	if err != nil {
		return nil, err
	}
	if info == nil {
		info = &InfoDict{Custom: make(map[string]string)}
	}

	for k, v := range updates {
		switch k {
		case "Author":
			info.Author = v
		case "Title":
			info.Title = v
		case "Subject":
			info.Subject = v
		case "Keywords":
			info.Keywords = v
		case "Creator":
			info.Creator = v
		case "Producer":
			info.Producer = v
		default:
			if info.Custom == nil {
				info.Custom = make(map[string]string)
			}
			info.Custom[k] = v
		}
	}

	if err := p.SetInfoDict(info); err != nil {
		return nil, err
	}

	return p.Bytes()
}

// Context returns the underlying pdfcpu context for advanced manipulation.
func (p *PDF) Context() types.Dict {
	if p.ctx == nil || p.ctx.RootDict == nil {
		return nil
	}
	return p.ctx.RootDict
}

// Raw returns the original bytes passed to Parse.
func (p *PDF) Raw() []byte {
	return p.raw
}

// ObjectCount returns the number of objects in the xref table.
func (p *PDF) ObjectCount() int {
	if p.ctx == nil || p.ctx.XRefTable == nil {
		return 0
	}
	return len(p.ctx.XRefTable.Table)
}

// CreateTestPDF returns a minimal valid PDF for testing.
func CreateTestPDF() []byte {
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	buf.WriteString("%\xe2\xe3\xcf\xd3\n") // binary marker
	buf.WriteString("1 0 obj\n<</Type/Catalog/Pages 2 0 R>>\nendobj\n")
	buf.WriteString("2 0 obj\n<</Type/Pages/Kids[3 0 R]/Count 1>>\nendobj\n")
	buf.WriteString("3 0 obj\n<</Type/Page/MediaBox[0 0 612 792]/Parent 2 0 R>>\nendobj\n")
	buf.WriteString("xref\n")
	buf.WriteString("0 4\n")
	buf.WriteString("0000000000 65535 f \n")
	buf.WriteString("0000000015 00000 n \n")
	buf.WriteString("0000000061 00000 n \n")
	buf.WriteString("0000000111 00000 n \n")
	buf.WriteString("trailer\n<</Size 4/Root 1 0 R>>\n")
	buf.WriteString("startxref\n177\n")
	buf.WriteString("%%EOF\n")
	return buf.Bytes()
}

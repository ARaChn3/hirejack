package pdf

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// InfoDict represents the Document Information Dictionary.
type InfoDict struct {
	Author   string
	Title    string
	Subject  string
	Keywords string
	Creator  string
	Producer string
	Custom   map[string]string
}

// XMPPacket represents XMP metadata.
type XMPPacket struct {
	Raw     []byte            // Raw XMP XML
	Fields  map[string]string // Parsed fields (dc:creator, etc.)
	Doctype string            // DOCTYPE for XXE
}

// GetInfoDict returns the Document Information Dictionary.
func (p *PDF) GetInfoDict() (*InfoDict, error) {
	xrt := p.ctx.XRefTable
	if xrt.Info == nil {
		return nil, nil
	}

	obj, err := xrt.Dereference(*xrt.Info)
	if err != nil {
		return nil, fmt.Errorf("pdf: info dict: %w", err)
	}
	if obj == nil {
		return nil, nil
	}

	d, ok := obj.(types.Dict)
	if !ok {
		return nil, fmt.Errorf("pdf: info dict: expected Dict, got %T", obj)
	}

	info := &InfoDict{Custom: make(map[string]string)}
	info.Author = stringFromDict(d, "Author")
	info.Title = stringFromDict(d, "Title")
	info.Subject = stringFromDict(d, "Subject")
	info.Keywords = stringFromDict(d, "Keywords")
	info.Creator = stringFromDict(d, "Creator")
	info.Producer = stringFromDict(d, "Producer")

	standardKeys := map[string]bool{
		"Author": true, "Title": true, "Subject": true,
		"Keywords": true, "Creator": true, "Producer": true,
		"CreationDate": true, "ModDate": true, "Trapped": true,
	}
	for k := range d {
		if !standardKeys[k] {
			if v := stringFromDict(d, k); v != "" {
				info.Custom[k] = v
			}
		}
	}

	return info, nil
}

// SetInfoDict updates the Document Information Dictionary.
func (p *PDF) SetInfoDict(info *InfoDict) error {
	xrt := p.ctx.XRefTable

	d := types.Dict{}
	if info.Author != "" {
		d["Author"] = types.StringLiteral(info.Author)
	}
	if info.Title != "" {
		d["Title"] = types.StringLiteral(info.Title)
	}
	if info.Subject != "" {
		d["Subject"] = types.StringLiteral(info.Subject)
	}
	if info.Keywords != "" {
		d["Keywords"] = types.StringLiteral(info.Keywords)
	}
	if info.Creator != "" {
		d["Creator"] = types.StringLiteral(info.Creator)
	}
	if info.Producer != "" {
		d["Producer"] = types.StringLiteral(info.Producer)
	}
	for k, v := range info.Custom {
		d[k] = types.StringLiteral(v)
	}

	if xrt.Info == nil {
		ir, err := xrt.IndRefForNewObject(d)
		if err != nil {
			return fmt.Errorf("pdf: set info: %w", err)
		}
		xrt.Info = ir
	} else {
		entry, ok := xrt.FindTableEntry(xrt.Info.ObjectNumber.Value(), xrt.Info.GenerationNumber.Value())
		if !ok {
			return fmt.Errorf("pdf: set info: entry not found")
		}
		entry.Object = d
	}

	return nil
}

// GetXMP returns the XMP metadata packet (nil if none).
func (p *PDF) GetXMP() (*XMPPacket, error) {
	xrt := p.ctx.XRefTable
	rootDict, err := xrt.Catalog()
	if err != nil {
		return nil, fmt.Errorf("pdf: xmp: catalog: %w", err)
	}

	metaRef := rootDict.IndirectRefEntry("Metadata")
	if metaRef == nil {
		return nil, nil
	}

	obj, err := xrt.Dereference(*metaRef)
	if err != nil {
		return nil, fmt.Errorf("pdf: xmp: deref: %w", err)
	}

	sd, ok := obj.(types.StreamDict)
	if !ok {
		return nil, fmt.Errorf("pdf: xmp: expected StreamDict, got %T", obj)
	}

	if err := sd.Decode(); err != nil {
		return nil, fmt.Errorf("pdf: xmp: decode: %w", err)
	}

	return &XMPPacket{
		Raw:    sd.Content,
		Fields: make(map[string]string),
	}, nil
}

// SetXMP sets or creates the XMP metadata.
func (p *PDF) SetXMP(xmp *XMPPacket) error {
	xrt := p.ctx.XRefTable
	rootDict, err := xrt.Catalog()
	if err != nil {
		return fmt.Errorf("pdf: set xmp: catalog: %w", err)
	}

	content := xmp.Raw
	if xmp.Doctype != "" {
		content = append([]byte(xmp.Doctype+"\n"), content...)
	}

	sd := types.StreamDict{
		Dict: types.Dict{
			"Type":    types.Name("Metadata"),
			"Subtype": types.Name("XML"),
		},
		Content: content,
	}
	if err := sd.Encode(); err != nil {
		return fmt.Errorf("pdf: set xmp: encode: %w", err)
	}

	ir, err := xrt.IndRefForNewObject(sd)
	if err != nil {
		return fmt.Errorf("pdf: set xmp: %w", err)
	}

	rootDict["Metadata"] = *ir
	return nil
}

// NewXMPPacket creates an empty XMP packet with basic structure.
func NewXMPPacket() *XMPPacket {
	return &XMPPacket{
		Raw: []byte(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description rdf:about=""
      xmlns:dc="http://purl.org/dc/elements/1.1/">
    </rdf:Description>
  </rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`),
		Fields: make(map[string]string),
	}
}

func stringFromDict(d types.Dict, key string) string {
	obj, ok := d[key]
	if !ok {
		return ""
	}
	switch v := obj.(type) {
	case types.StringLiteral:
		return v.Value()
	case types.HexLiteral:
		return v.Value()
	default:
		return ""
	}
}

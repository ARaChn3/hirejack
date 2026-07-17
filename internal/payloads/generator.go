package payloads

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Generator creates payload content procedurally from components.
type Generator struct {
	rng *rand.Rand
}

// NewGenerator creates a payload generator with the given seed.
// If seed is 0, uses current time.
func NewGenerator(seed int64) *Generator {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &Generator{rng: rand.New(rand.NewSource(seed))}
}

func (g *Generator) pick(opts []string) string {
	if len(opts) == 0 {
		return ""
	}
	return opts[g.rng.Intn(len(opts))]
}

func (g *Generator) pickN(opts []string, n int) []string {
	if n >= len(opts) {
		return opts
	}
	cp := make([]string, len(opts))
	copy(cp, opts)
	g.rng.Shuffle(len(cp), func(i, j int) { cp[i], cp[j] = cp[j], cp[i] })
	return cp[:n]
}

// GenerateLLMInjection creates a prompt injection payload.
func (g *Generator) GenerateLLMInjection(variant string) string {
	switch variant {
	case "direct", "ignore_resume", "system_override":
		return g.pick(LLMInjectionPayloads[:6])
	case "roleplay", "helpful_assistant":
		return g.pick(LLMInjectionPayloads[6:9])
	case "delimiter", "delimiter_escape", "markdown_escape":
		return g.pick(LLMInjectionPayloads[9:12])
	case "output_format", "json_inject", "structured_output":
		return g.pick(LLMInjectionPayloads[12:15])
	case "exfil", "prompt_leak", "instruction_dump":
		return g.pick(LLMInjectionPayloads[15:19])
	case "jailbreak", "hypothetical", "debug_mode":
		return g.pick(LLMInjectionPayloads[19:22])
	case "multi_stage", "chained":
		return g.pick(LLMInjectionPayloads[22:23])
	case "persona", "character":
		return g.pick(LLMInjectionPayloads[23:24])
	case "context_poison", "social_proof":
		return g.pick(LLMInjectionPayloads[24:])
	default:
		return g.pick(LLMInjectionPayloads)
	}
}

// GenerateSQLi creates a SQL injection payload.
func (g *Generator) GenerateSQLi(variant string) string {
	switch variant {
	case "basic", "name_inject", "email_inject":
		return g.pick(SQLiPayloads[:8])
	case "blind", "blind_mysql", "blind_postgres", "blind_mssql":
		return g.pick(SQLiPayloads[8:14])
	case "union", "union_exfil":
		return g.pick(SQLiPayloads[14:19])
	case "stacked", "stacked_queries":
		return g.pick(SQLiPayloads[19:24])
	case "mysql":
		return g.pick(SQLiPayloads[24:27])
	case "postgres":
		return g.pick(SQLiPayloads[27:29])
	case "mssql":
		return g.pick(SQLiPayloads[29:])
	default:
		return g.pick(SQLiPayloads)
	}
}

// GenerateXXE creates an XXE payload.
func (g *Generator) GenerateXXE(variant string) string {
	switch variant {
	case "file", "dtd_external", "file_disclosure":
		return g.pick(XXEPayloads[:6])
	case "php", "php_filter":
		return XXEPayloads[6]
	case "ssrf", "aws_metadata", "gcp_metadata", "cloud_ssrf":
		return g.pick(XXEPayloads[7:10])
	case "oob", "oob_exfil", "param_entity":
		return g.pick(XXEPayloads[10:13])
	case "bomb", "billion_laughs":
		return XXEPayloads[13]
	default:
		return g.pick(XXEPayloads)
	}
}

// GenerateDoS creates a denial-of-service payload.
func (g *Generator) GenerateDoS(variant string) string {
	switch variant {
	case "billion_laughs", "xml_bomb":
		return DoSPayloads[0]
	case "quadratic_blowup", "quadratic":
		return DoSPayloads[1]
	case "redos", "regex":
		return g.pick(DoSPayloads[2:4])
	case "deep_nesting", "recursion":
		return DoSPayloads[4]
	case "long_string", "bloat", "field_overflow":
		return DoSPayloads[5]
	default:
		return g.pick(DoSPayloads)
	}
}

// GenerateIntegrity creates an integrity/evasion payload.
func (g *Generator) GenerateIntegrity(variant string) string {
	switch variant {
	case "homoglyph", "homoglyph_company", "unicode_confusables":
		return g.pick(IntegrityPayloads[:3])
	case "bidi", "bidi_injection", "rtl":
		return g.pick(IntegrityPayloads[3:5])
	case "zwc", "zwc_hidden", "zero_width":
		return g.pick(IntegrityPayloads[5:7])
	case "null_byte", "null":
		return g.pick(IntegrityPayloads[7:9])
	case "crlf", "crlf_injection", "header_inject":
		return IntegrityPayloads[9]
	case "normalization", "ligature":
		return g.pick(IntegrityPayloads[10:12])
	case "truncation", "truncation_attack", "overflow":
		return g.pick(IntegrityPayloads[12:])
	default:
		return g.pick(IntegrityPayloads)
	}
}

// GenerateMLPoison creates an ML poisoning payload.
func (g *Generator) GenerateMLPoison(variant string) string {
	switch variant {
	case "white_keywords", "invisible_text":
		return MLPoisonPayloads[0]
	case "metadata_keywords", "keyword_stuff":
		return MLPoisonPayloads[1]
	case "unicode_keywords", "invisible_unicode":
		return MLPoisonPayloads[2]
	case "micro_text", "font_zero":
		return MLPoisonPayloads[3]
	case "off_canvas", "position_hidden":
		return MLPoisonPayloads[4]
	case "embedding_anchor", "semantic_poison":
		return g.pick(MLPoisonPayloads[5:])
	default:
		return g.pick(MLPoisonPayloads)
	}
}

// Keyword stuffing components
var techSkills = []string{
	"Python", "JavaScript", "TypeScript", "Go", "Rust", "Java", "C++",
	"React", "Vue", "Angular", "Node.js", "Django", "FastAPI",
	"AWS", "GCP", "Azure", "Kubernetes", "Docker", "Terraform",
	"PostgreSQL", "MongoDB", "Redis", "Elasticsearch",
	"Machine Learning", "Deep Learning", "NLP", "Computer Vision",
	"Agile", "Scrum", "CI/CD", "DevOps", "SRE",
}

// GenerateKeywordStuffing creates invisible keyword content.
func (g *Generator) GenerateKeywordStuffing(count int) string {
	skills := g.pickN(techSkills, count)
	return strings.Join(skills, " ")
}

// Homoglyph mappings for integrity attacks
var homoglyphs = map[rune][]rune{
	'a': {'а', 'ɑ', 'α'},
	'e': {'е', 'ε', 'ℯ'},
	'o': {'о', 'ο', '0', 'ⲟ'},
	'i': {'і', 'ι', '1', 'ⅰ'},
	'c': {'с', 'ϲ', 'ⅽ'},
	'p': {'р', 'ρ'},
	's': {'ѕ', 'ꜱ'},
	'x': {'х', 'ⅹ'},
}

// GenerateHomoglyph replaces characters with visual lookalikes.
func (g *Generator) GenerateHomoglyph(text string, rate float64) string {
	var result strings.Builder
	for _, r := range text {
		if alts, ok := homoglyphs[r]; ok && g.rng.Float64() < rate {
			result.WriteRune(alts[g.rng.Intn(len(alts))])
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// FillTemplate generates content for a payload by category and ID.
func (g *Generator) FillTemplate(payloadID string) string {
	parts := strings.SplitN(payloadID, ".", 2)
	if len(parts) != 2 {
		return ""
	}
	category, variant := parts[0], parts[1]

	switch category {
	case "llm_injection":
		return g.GenerateLLMInjection(variant)
	case "sqli":
		return g.GenerateSQLi(variant)
	case "xxe":
		return g.GenerateXXE(variant)
	case "dos":
		return g.GenerateDoS(variant)
	case "integrity":
		return g.GenerateIntegrity(variant)
	case "ml_poison":
		return g.GenerateMLPoison(variant)
	default:
		return ""
	}
}

// GenerateNLPOverload creates a very long sentence to overload NLP parsers.
func (g *Generator) GenerateNLPOverload(words int) string {
	fillers := []string{
		"experienced", "professional", "skilled", "expert", "senior",
		"with", "and", "including", "such as", "for example",
		"developing", "implementing", "designing", "managing", "leading",
		"software", "systems", "applications", "solutions", "platforms",
	}
	var parts []string
	for i := 0; i < words; i++ {
		parts = append(parts, g.pick(fillers))
	}
	return strings.Join(parts, " ")
}

// GenerateOCRBomb creates text designed to confuse OCR.
func (g *Generator) GenerateOCRBomb() string {
	confusing := []string{
		"l1I|", "O0Q", "S5$", "Z2", "B8", "G6", "rn=m", "vv=w", "cl=d",
	}
	var result strings.Builder
	for i := 0; i < 100; i++ {
		result.WriteString(g.pick(confusing))
	}
	return result.String()
}

// GenerateFontBomb creates invalid font references.
func (g *Generator) GenerateFontBomb() string {
	return fmt.Sprintf(`/F1 %d Tf (%s) Tj`, g.rng.Intn(9999), strings.Repeat("A", 10000))
}

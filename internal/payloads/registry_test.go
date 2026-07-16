package payloads

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

func TestRegistryCategoriesDefined(t *testing.T) {
	want := []string{"sqli", "xxe", "dos", "integrity", "ml_poison"}
	if len(Registry) != len(want) {
		t.Fatalf("expected %d categories, got %d", len(want), len(Registry))
	}
	for _, id := range want {
		cat := GetCategory(id)
		if cat == nil {
			t.Fatalf("expected category %q to be defined", id)
		}
		if cat.Name == "" {
			t.Errorf("category %q has empty Name", id)
		}
		if cat.Description == "" {
			t.Errorf("category %q has empty Description", id)
		}
		if cat.Color == "" {
			t.Errorf("category %q has empty Color", id)
		}
		if len(cat.Payloads) == 0 {
			t.Errorf("category %q has no payloads", id)
		}
	}
}

func TestCategoryColors(t *testing.T) {
	want := map[string]string{
		"sqli":      "warn",
		"xxe":       "danger",
		"dos":       "danger",
		"integrity": "info",
		"ml_poison": "accent",
	}
	for id, color := range want {
		cat := GetCategory(id)
		if cat == nil {
			t.Fatalf("category %q not found", id)
		}
		if cat.Color != color {
			t.Errorf("category %q: expected color %q, got %q", id, color, cat.Color)
		}
	}
}

func TestGetCategoryUnknown(t *testing.T) {
	if cat := GetCategory("does_not_exist"); cat != nil {
		t.Errorf("expected nil for unknown category, got %+v", cat)
	}
}

func TestGetPayloadLookup(t *testing.T) {
	cases := []struct {
		id       string
		wantName string
	}{
		{"sqli.basic", "Basic Drop Table"},
		{"sqli.blind", "Blind Time-Based"},
		{"sqli.union", "Union-Based Extraction"},
		{"xxe.file_read", "Local File Read"},
		{"xxe.ssrf", "Cloud Metadata SSRF"},
		{"xxe.callback", "Out-of-Band Callback"},
		{"xxe.billion_laughs", "Billion Laughs"},
		{"dos.entity_expansion", "Entity Expansion"},
		{"dos.recursion", "Circular Xref"},
		{"dos.bloat", "Orphaned Object Bloat"},
		{"integrity.unicode", "Unicode Tricks"},
		{"integrity.truncation", "Field Truncation"},
		{"integrity.encoding", "Mixed Encoding"},
		{"ml_poison.skills", "Hidden Skills"},
		{"ml_poison.titles", "Hidden Job Titles"},
		{"ml_poison.keywords", "Keyword Stuffing"},
	}
	for _, c := range cases {
		p := GetPayload(c.id)
		if p == nil {
			t.Fatalf("expected payload %q to be found", c.id)
		}
		if p.Name != c.wantName {
			t.Errorf("payload %q: expected Name %q, got %q", c.id, c.wantName, p.Name)
		}
	}
}

func TestGetPayloadUnknownOrMalformed(t *testing.T) {
	cases := []string{
		"sqli.nonexistent",
		"nonexistent.basic",
		"malformed_no_dot",
		"",
	}
	for _, id := range cases {
		if p := GetPayload(id); p != nil {
			t.Errorf("expected nil for id %q, got %+v", id, p)
		}
	}
}

func TestSQLiTemplates(t *testing.T) {
	cases := map[string]string{
		"sqli.basic": `'; DROP TABLE {{table}}; --`,
		"sqli.blind": `'; WAITFOR DELAY '0:0:5'; --`,
		"sqli.union": `' UNION SELECT password FROM users --`,
	}
	for id, want := range cases {
		p := GetPayload(id)
		if p == nil {
			t.Fatalf("payload %q not found", id)
		}
		if p.Template != want {
			t.Errorf("payload %q: expected template %q, got %q", id, want, p.Template)
		}
	}
}

func TestImpactValuesAreValid(t *testing.T) {
	valid := map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
	for _, cat := range Registry {
		for _, p := range cat.Payloads {
			if !valid[p.Impact] {
				t.Errorf("payload %s.%s has invalid impact %q", cat.ID, p.ID, p.Impact)
			}
		}
	}
}

func TestProfileIDs(t *testing.T) {
	want := map[string]bool{
		"sqli_focus":    true,
		"xxe_focus":     true,
		"dos_focus":     true,
		"full_audit":    true,
		"stealth_recon": true,
	}
	got := ProfileIDs()
	if len(got) != len(want) {
		t.Fatalf("expected %d profiles, got %d: %v", len(want), len(got), got)
	}
	for _, id := range got {
		if !want[id] {
			t.Errorf("unexpected profile ID %q", id)
		}
	}
}

func TestGetProfileSQLiFocus(t *testing.T) {
	got := GetProfile("sqli_focus")
	want := []string{"sqli.basic", "sqli.blind", "sqli.union"}
	assertSameIDs(t, got, want)
}

func TestGetProfileXXEFocus(t *testing.T) {
	got := GetProfile("xxe_focus")
	want := []string{"xxe.file_read", "xxe.ssrf", "xxe.callback", "xxe.billion_laughs"}
	assertSameIDs(t, got, want)
}

func TestGetProfileDoSFocus(t *testing.T) {
	got := GetProfile("dos_focus")
	want := []string{"dos.entity_expansion", "dos.recursion", "dos.bloat"}
	assertSameIDs(t, got, want)
}

func TestGetProfileFullAudit(t *testing.T) {
	got := GetProfile("full_audit")
	var want []string
	for _, cat := range Registry {
		for _, p := range cat.Payloads {
			want = append(want, cat.ID+"."+p.ID)
		}
	}
	assertSameIDs(t, got, want)
}

func TestGetProfileStealthRecon(t *testing.T) {
	got := GetProfile("stealth_recon")
	if len(got) == 0 {
		t.Fatal("expected stealth_recon profile to be non-empty")
	}
	for _, id := range got {
		p := GetPayload(id)
		if p == nil {
			t.Fatalf("stealth_recon referenced unknown payload %q", id)
		}
		hasLevel := false
		for _, s := range p.Stealth {
			if s == int(LevelShadowed) || s == int(LevelPolyglot) {
				hasLevel = true
				break
			}
		}
		if !hasLevel {
			t.Errorf("payload %q in stealth_recon lacks level 4/5 compatibility: %v", id, p.Stealth)
		}
	}
}

func TestGetProfileUnknown(t *testing.T) {
	if got := GetProfile("does_not_exist"); got != nil {
		t.Errorf("expected nil for unknown profile, got %v", got)
	}
}

func TestPayloadIDsUniqueWithinCategory(t *testing.T) {
	for _, cat := range Registry {
		seen := map[string]bool{}
		for _, p := range cat.Payloads {
			if seen[p.ID] {
				t.Errorf("duplicate payload ID %q in category %q", p.ID, cat.ID)
			}
			seen[p.ID] = true
		}
	}
}

func TestTemplateNotExposedInJSON(t *testing.T) {
	p := GetPayload("sqli.basic")
	if p == nil {
		t.Fatal("sqli.basic not found")
	}
	if p.Template == "" {
		t.Fatal("expected sqli.basic to have a template")
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	if strings.Contains(string(b), "DROP TABLE") {
		t.Errorf("Template contents leaked into JSON output: %s", b)
	}
	if strings.Contains(string(b), `"Template"`) || strings.Contains(string(b), `"template"`) {
		t.Errorf("Template field key leaked into JSON output: %s", b)
	}
}

func assertSameIDs(t *testing.T, got, want []string) {
	t.Helper()
	gotSorted := append([]string(nil), got...)
	wantSorted := append([]string(nil), want...)
	sort.Strings(gotSorted)
	sort.Strings(wantSorted)
	if len(gotSorted) != len(wantSorted) {
		t.Fatalf("expected %v, got %v", wantSorted, gotSorted)
	}
	for i := range gotSorted {
		if gotSorted[i] != wantSorted[i] {
			t.Fatalf("expected %v, got %v", wantSorted, gotSorted)
		}
	}
}

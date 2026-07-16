package generators

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	g := NewPersonaGenerator()
	p := g.Generate()

	if p.FirstName == "" {
		t.Error("FirstName should not be empty")
	}
	if p.LastName == "" {
		t.Error("LastName should not be empty")
	}
	if p.Email == "" {
		t.Error("Email should not be empty")
	}
	if !strings.Contains(p.Email, "@") {
		t.Error("Email should contain @")
	}
}

func TestFullName(t *testing.T) {
	p := &Persona{FirstName: "John", LastName: "Doe"}
	if p.FullName() != "John Doe" {
		t.Errorf("FullName = %q, want %q", p.FullName(), "John Doe")
	}
}

func TestGenerateUnique(t *testing.T) {
	g := NewPersonaGenerator()

	emails := make(map[string]bool)
	for i := 0; i < 10; i++ {
		p := g.Generate()
		// Not strictly unique but should vary
		emails[p.Email] = true
	}

	// Should have some variation (not all identical)
	if len(emails) < 2 {
		t.Error("personas should have variation")
	}
}

func TestGeneratePhone(t *testing.T) {
	g := NewPersonaGenerator()
	p := g.Generate()

	if !strings.Contains(p.Phone, "(") || !strings.Contains(p.Phone, ")") {
		t.Error("Phone should be formatted with parentheses")
	}
}

func TestGenerateExperience(t *testing.T) {
	g := NewPersonaGenerator()
	p := g.Generate()

	if len(p.Experience) < 2 {
		t.Error("should have at least 2 jobs")
	}

	for _, job := range p.Experience {
		if job.Company == "" {
			t.Error("job company should not be empty")
		}
		if job.Title == "" {
			t.Error("job title should not be empty")
		}
	}
}

func TestGenerateEducation(t *testing.T) {
	g := NewPersonaGenerator()
	p := g.Generate()

	if len(p.Education) == 0 {
		t.Error("should have education")
	}

	for _, school := range p.Education {
		if school.Name == "" {
			t.Error("school name should not be empty")
		}
		if school.Degree == "" {
			t.Error("degree should not be empty")
		}
	}
}

func TestGenerateSkills(t *testing.T) {
	g := NewPersonaGenerator()
	p := g.Generate()

	if len(p.Skills) == 0 {
		t.Error("should have skills")
	}

	// Skills should be unique
	seen := make(map[string]bool)
	for _, skill := range p.Skills {
		if seen[skill] {
			t.Errorf("duplicate skill: %s", skill)
		}
		seen[skill] = true
	}
}

func TestGenerateSummary(t *testing.T) {
	g := NewPersonaGenerator()
	p := g.Generate()

	if p.Summary == "" {
		t.Error("summary should not be empty")
	}
	if !strings.Contains(p.Summary, "years") {
		t.Error("summary should mention years of experience")
	}
}

func TestGenerateAddress(t *testing.T) {
	g := NewPersonaGenerator()
	p := g.Generate()

	if p.Address == "" {
		t.Error("address should not be empty")
	}
	if p.City == "" {
		t.Error("city should not be empty")
	}
}

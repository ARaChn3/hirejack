package engine

import (
	"testing"

	"github.com/novusedge/hirejack/internal/pdf"
)

func TestNew(t *testing.T) {
	e := New("https://example.com")
	if e == nil {
		t.Fatal("New returned nil")
	}
	if e.canaryStore == nil {
		t.Error("canaryStore should not be nil")
	}
}

func TestInjectBasic(t *testing.T) {
	e := New("https://example.com")
	e.SetStealth(StealthBasic)

	result, err := e.Inject(InjectRequest{
		PDF:        pdf.CreateTestPDF(),
		CustomData: map[string]string{"Author": "test payload"},
	})
	if err != nil {
		t.Fatalf("Inject: %v", err)
	}

	if len(result.PDF) == 0 {
		t.Error("PDF should not be empty")
	}
	if len(result.Injections) == 0 {
		t.Error("should have injections")
	}
}

func TestInjectWithStealth(t *testing.T) {
	e := New("https://example.com")

	levels := []StealthLevel{
		StealthBasic,
		StealthHex,
		StealthSplit,
		StealthOrphan,
		StealthMax,
	}

	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			result, err := e.Inject(InjectRequest{
				PDF:        pdf.CreateTestPDF(),
				CustomData: map[string]string{"Test": "value"},
				Stealth:    level,
			})
			if err != nil {
				t.Errorf("Inject at level %d: %v", level, err)
			}
			if len(result.PDF) == 0 {
				t.Error("PDF should not be empty")
			}
		})
	}
}

func TestInjectWithCanary(t *testing.T) {
	e := New("https://audit.example.com")

	result, err := e.Inject(InjectRequest{
		PDF:         pdf.CreateTestPDF(),
		AddCanary:   true,
		CanaryLabel: "test-target",
	})
	if err != nil {
		t.Fatalf("Inject: %v", err)
	}

	if result.CanaryToken == nil {
		t.Error("should have canary token")
	}
	if result.CanaryToken.Target != "test-target" {
		t.Errorf("Target = %q", result.CanaryToken.Target)
	}

	// Token should be in store
	stored := e.CanaryStore().Get(result.CanaryToken.ID)
	if stored == nil {
		t.Error("token should be in store")
	}
}

func TestGenerateCV(t *testing.T) {
	e := New("https://example.com")

	result, err := e.GenerateCV(nil, StealthBasic)
	if err != nil {
		t.Fatalf("GenerateCV: %v", err)
	}

	if len(result.PDF) == 0 {
		t.Error("PDF should not be empty")
	}

	// Should be valid PDF
	_, err = pdf.Parse(result.PDF)
	if err != nil {
		t.Errorf("reparse: %v", err)
	}
}

func TestStealthLevelString(t *testing.T) {
	tests := []struct {
		level StealthLevel
		want  string
	}{
		{StealthNone, "none"},
		{StealthBasic, "basic"},
		{StealthHex, "hex"},
		{StealthSplit, "split"},
		{StealthOrphan, "orphan"},
		{StealthMax, "max"},
	}

	for _, tc := range tests {
		if tc.level.String() != tc.want {
			t.Errorf("%d.String() = %q, want %q", tc.level, tc.level.String(), tc.want)
		}
	}
}

// String returns a human-readable stealth level name.
func (s StealthLevel) String() string {
	switch s {
	case StealthNone:
		return "none"
	case StealthBasic:
		return "basic"
	case StealthHex:
		return "hex"
	case StealthSplit:
		return "split"
	case StealthOrphan:
		return "orphan"
	case StealthMax:
		return "max"
	default:
		return "unknown"
	}
}

func TestInjectMultiplePayloads(t *testing.T) {
	e := New("https://example.com")

	result, err := e.Inject(InjectRequest{
		PDF: pdf.CreateTestPDF(),
		CustomData: map[string]string{
			"Author":  "payload1",
			"Title":   "payload2",
			"Subject": "payload3",
		},
		Stealth: StealthMax,
	})
	if err != nil {
		t.Fatalf("Inject: %v", err)
	}

	if len(result.Injections) < 3 {
		t.Errorf("should have at least 3 injections, got %d", len(result.Injections))
	}
}

func TestCanaryStore(t *testing.T) {
	e := New("https://example.com")

	// Generate some canaries via injections
	e.Inject(InjectRequest{
		PDF:       pdf.CreateTestPDF(),
		AddCanary: true,
	})
	e.Inject(InjectRequest{
		PDF:       pdf.CreateTestPDF(),
		AddCanary: true,
	})

	store := e.CanaryStore()
	if store.Count() != 2 {
		t.Errorf("store should have 2 tokens, got %d", store.Count())
	}
}

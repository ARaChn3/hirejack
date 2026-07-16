package encoders

import (
	"strings"
	"testing"
)

func TestEncodeName(t *testing.T) {
	e := NewHexEncoder()

	result := e.EncodeName("Author")
	if !strings.Contains(result, "#") {
		t.Error("should contain hex escapes")
	}

	decoded := DecodePDFName(result)
	if decoded != "Author" {
		t.Errorf("decode mismatch: got %q, want %q", decoded, "Author")
	}
}

func TestEncodeXMLEntities(t *testing.T) {
	e := NewHexEncoder()

	payload := "'; DROP TABLE users; --"
	result := e.EncodeXMLEntities(payload)

	if strings.Contains(result, "'") {
		t.Error("should not contain raw quote")
	}
	if !strings.Contains(result, "&#39;") {
		t.Error("should contain XML entity for quote")
	}
}

func TestRoundTrip(t *testing.T) {
	e := NewHexEncoder()

	original := "test payload"
	encoded := e.Encode(original)

	if !strings.Contains(encoded, "#") {
		t.Error("encoded should contain hex escapes")
	}

	decoded := DecodeHex(encoded)
	if decoded != original {
		t.Errorf("decode mismatch: got %q, want %q", decoded, original)
	}
}

func TestEncodeAll(t *testing.T) {
	e := NewHexEncoder()

	result := e.EncodeAll("ABC")
	if result != "#41#42#43" {
		t.Errorf("EncodeAll(ABC) = %q, want #41#42#43", result)
	}

	decoded := DecodeHex(result)
	if decoded != "ABC" {
		t.Errorf("decode mismatch: got %q", decoded)
	}
}

func TestWithRatio(t *testing.T) {
	// Ratio 1 = encode every char
	e := NewHexEncoder().WithRatio(1)
	result := e.Encode("AB")
	if result != "#41#42" {
		t.Errorf("ratio 1: got %q, want #41#42", result)
	}

	// Ratio 2 = encode every other char
	e = NewHexEncoder().WithRatio(2)
	result = e.Encode("ABCD")
	if result != "#41B#43D" {
		t.Errorf("ratio 2: got %q, want #41B#43D", result)
	}
}

func TestDecodeHex(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"#41#42#43", "ABC"},
		{"A#42C", "ABC"},
		{"ABC", "ABC"},
		{"#00", "\x00"},
		{"#FF", "\xFF"},
		{"#4", "#4"},        // incomplete, not decoded
		{"##41", "#A"},      // # then #41
	}

	for _, tc := range tests {
		got := DecodeHex(tc.input)
		if got != tc.expected {
			t.Errorf("DecodeHex(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestEncodeNameSpecialChars(t *testing.T) {
	e := NewHexEncoder()

	// Special chars should always be encoded
	result := e.EncodeName("A/B")
	if !strings.Contains(result, "#2F") {
		t.Error("/ should be encoded as #2F")
	}

	result = e.EncodeName("A#B")
	if !strings.Contains(result, "#23") {
		t.Error("# should be encoded as #23")
	}
}

func TestXMLEntitiesAll(t *testing.T) {
	e := NewHexEncoder()

	tests := []struct {
		char     string
		expected string
	}{
		{"'", "&#39;"},
		{"\"", "&#34;"},
		{"<", "&#60;"},
		{">", "&#62;"},
		{"&", "&#38;"},
	}

	for _, tc := range tests {
		got := e.EncodeXMLEntities(tc.char)
		if got != tc.expected {
			t.Errorf("EncodeXMLEntities(%q) = %q, want %q", tc.char, got, tc.expected)
		}
	}
}

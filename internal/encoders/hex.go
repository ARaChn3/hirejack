// Package encoders provides stealth encoding for payloads.
package encoders

import (
	"fmt"
	"strconv"
	"strings"
)

// Encoder interface shared across encoders.
type Encoder interface {
	Encode(payload string) string
	EncodeName(name string) string
}

// HexEncoder provides Level 2 stealth encoding.
type HexEncoder struct {
	ratio int // encode every Nth char (3 = ~33%)
}

// NewHexEncoder creates a new hex encoder.
// Default encodes ~33% of characters.
func NewHexEncoder() *HexEncoder {
	return &HexEncoder{ratio: 3}
}

// WithRatio sets encoding density (lower = more encoding).
func (e *HexEncoder) WithRatio(n int) *HexEncoder {
	if n < 1 {
		n = 1
	}
	e.ratio = n
	return e
}

// Encode applies hex encoding to payload string.
// Encodes some characters as #XX hex escapes.
func (e *HexEncoder) Encode(payload string) string {
	var result strings.Builder
	for i, r := range payload {
		if i%e.ratio == 0 && r < 128 {
			result.WriteString(fmt.Sprintf("#%02X", r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// EncodeName hex-escapes a PDF name.
// Example: Author → Au#74hor
func (e *HexEncoder) EncodeName(name string) string {
	var result strings.Builder
	for i, r := range name {
		if shouldEncodeChar(r) || (isAlnum(r) && i%e.ratio == 0) {
			result.WriteString(fmt.Sprintf("#%02X", r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// EncodeXMLEntities encodes for XMP context.
// Uses XML numeric character references: &#39; for '
func (e *HexEncoder) EncodeXMLEntities(payload string) string {
	replacer := strings.NewReplacer(
		"'", "&#39;",
		"\"", "&#34;",
		"<", "&#60;",
		">", "&#62;",
		"&", "&#38;",
	)
	return replacer.Replace(payload)
}

// EncodeAll encodes every character as hex.
func (e *HexEncoder) EncodeAll(payload string) string {
	var result strings.Builder
	for _, r := range payload {
		if r < 128 {
			result.WriteString(fmt.Sprintf("#%02X", r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// DecodeHex decodes hex-escaped strings (#XX format).
func DecodeHex(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '#' && i+2 < len(s) {
			hex := s[i+1 : i+3]
			if b, err := strconv.ParseUint(hex, 16, 8); err == nil {
				result.WriteByte(byte(b))
				i += 3
				continue
			}
		}
		result.WriteByte(s[i])
		i++
	}
	return result.String()
}

// DecodePDFName decodes hex escapes in PDF names.
func DecodePDFName(s string) string {
	return DecodeHex(s)
}

func shouldEncodeChar(r rune) bool {
	// PDF spec special chars that should be encoded
	return r == '#' || r == '/' || r == '(' || r == ')' ||
		r == '<' || r == '>' || r == '[' || r == ']' ||
		r == '{' || r == '}' || r == '%' || r <= 32 || r >= 127
}

func isAlnum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

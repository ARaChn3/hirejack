package encoders

import (
	"crypto/rand"
	"encoding/hex"
)

// SplitEncoder provides Level 3 stealth encoding.
// Splits payloads across multiple fragments with reassembly markers.
type SplitEncoder struct {
	chunkSize int
	markerLen int
}

// Fragment represents a split piece of a payload.
type Fragment struct {
	ID      string `json:"id"`      // Unique fragment ID
	Index   int    `json:"index"`   // Order in sequence
	Total   int    `json:"total"`   // Total fragments
	Content []byte `json:"content"` // Fragment data
}

// SplitResult contains all fragments and reassembly info.
type SplitResult struct {
	PayloadID  string     `json:"payload_id"` // Shared ID across fragments
	Fragments  []Fragment `json:"fragments"`
	Order      []string   `json:"order"` // Fragment IDs in order
}

// NewSplitEncoder creates a new split encoder.
// Default chunk size is 64 bytes.
func NewSplitEncoder() *SplitEncoder {
	return &SplitEncoder{
		chunkSize: 64,
		markerLen: 8,
	}
}

// WithChunkSize sets the fragment size.
func (e *SplitEncoder) WithChunkSize(n int) *SplitEncoder {
	if n < 8 {
		n = 8
	}
	e.chunkSize = n
	return e
}

// Split divides a payload into fragments.
func (e *SplitEncoder) Split(payload []byte) (*SplitResult, error) {
	payloadID := e.generateID()

	chunks := e.chunk(payload)
	fragments := make([]Fragment, len(chunks))
	order := make([]string, len(chunks))

	for i, chunk := range chunks {
		fragID := e.generateID()
		fragments[i] = Fragment{
			ID:      fragID,
			Index:   i,
			Total:   len(chunks),
			Content: chunk,
		}
		order[i] = fragID
	}

	return &SplitResult{
		PayloadID: payloadID,
		Fragments: fragments,
		Order:     order,
	}, nil
}

// Reassemble combines fragments back into the original payload.
func (e *SplitEncoder) Reassemble(result *SplitResult) []byte {
	// Build order map
	orderMap := make(map[string]int)
	for i, id := range result.Order {
		orderMap[id] = i
	}

	// Sort fragments by order
	sorted := make([]Fragment, len(result.Fragments))
	for _, frag := range result.Fragments {
		if idx, ok := orderMap[frag.ID]; ok {
			sorted[idx] = frag
		}
	}

	// Concatenate
	var payload []byte
	for _, frag := range sorted {
		payload = append(payload, frag.Content...)
	}
	return payload
}

// SplitString is a convenience method for string payloads.
func (e *SplitEncoder) SplitString(payload string) (*SplitResult, error) {
	return e.Split([]byte(payload))
}

// ReassembleString returns the reassembled payload as a string.
func (e *SplitEncoder) ReassembleString(result *SplitResult) string {
	return string(e.Reassemble(result))
}

func (e *SplitEncoder) chunk(data []byte) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(data); i += e.chunkSize {
		end := i + e.chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := make([]byte, end-i)
		copy(chunk, data[i:end])
		chunks = append(chunks, chunk)
	}
	return chunks
}

func (e *SplitEncoder) generateID() string {
	b := make([]byte, e.markerLen/2)
	rand.Read(b)
	return hex.EncodeToString(b)
}

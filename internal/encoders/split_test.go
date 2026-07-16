package encoders

import (
	"bytes"
	"testing"
)

func TestSplit(t *testing.T) {
	e := NewSplitEncoder().WithChunkSize(16)

	payload := []byte("This is a test payload that should be split into multiple chunks.")
	result, err := e.Split(payload)
	if err != nil {
		t.Fatalf("Split: %v", err)
	}

	if len(result.Fragments) < 2 {
		t.Error("should split into multiple fragments")
	}

	if result.PayloadID == "" {
		t.Error("PayloadID should be set")
	}

	if len(result.Order) != len(result.Fragments) {
		t.Error("Order should match Fragments count")
	}
}

func TestReassemble(t *testing.T) {
	e := NewSplitEncoder().WithChunkSize(16)

	original := []byte("Test payload for reassembly verification.")
	result, _ := e.Split(original)

	reassembled := e.Reassemble(result)
	if !bytes.Equal(original, reassembled) {
		t.Errorf("reassembled mismatch:\ngot:  %q\nwant: %q", reassembled, original)
	}
}

func TestSplitString(t *testing.T) {
	e := NewSplitEncoder().WithChunkSize(8)

	original := "Hello, World!"
	result, _ := e.SplitString(original)

	reassembled := e.ReassembleString(result)
	if reassembled != original {
		t.Errorf("got %q, want %q", reassembled, original)
	}
}

func TestFragmentIDs(t *testing.T) {
	e := NewSplitEncoder().WithChunkSize(8)

	result, _ := e.Split([]byte("test data for id check"))

	// All IDs should be unique
	seen := make(map[string]bool)
	for _, frag := range result.Fragments {
		if seen[frag.ID] {
			t.Errorf("duplicate fragment ID: %s", frag.ID)
		}
		seen[frag.ID] = true
	}
}

func TestFragmentMetadata(t *testing.T) {
	e := NewSplitEncoder().WithChunkSize(10)

	result, _ := e.Split([]byte("ABCDEFGHIJ1234567890")) // 20 bytes = 2 chunks

	if len(result.Fragments) != 2 {
		t.Fatalf("expected 2 fragments, got %d", len(result.Fragments))
	}

	for i, frag := range result.Fragments {
		if frag.Index != i {
			t.Errorf("fragment %d: Index = %d, want %d", i, frag.Index, i)
		}
		if frag.Total != 2 {
			t.Errorf("fragment %d: Total = %d, want 2", i, frag.Total)
		}
	}
}

func TestSmallPayload(t *testing.T) {
	e := NewSplitEncoder().WithChunkSize(100)

	// Payload smaller than chunk size
	original := []byte("tiny")
	result, _ := e.Split(original)

	if len(result.Fragments) != 1 {
		t.Errorf("expected 1 fragment for small payload, got %d", len(result.Fragments))
	}

	reassembled := e.Reassemble(result)
	if !bytes.Equal(original, reassembled) {
		t.Error("small payload reassembly failed")
	}
}

func TestEmptyPayload(t *testing.T) {
	e := NewSplitEncoder()

	result, _ := e.Split([]byte{})
	if len(result.Fragments) != 0 {
		t.Errorf("empty payload should produce 0 fragments, got %d", len(result.Fragments))
	}

	reassembled := e.Reassemble(result)
	if len(reassembled) != 0 {
		t.Error("empty reassembly should be empty")
	}
}

func TestLargePayload(t *testing.T) {
	e := NewSplitEncoder().WithChunkSize(64)

	// 1KB payload
	original := make([]byte, 1024)
	for i := range original {
		original[i] = byte(i % 256)
	}

	result, _ := e.Split(original)

	expectedChunks := (1024 + 63) / 64 // ceil(1024/64) = 16
	if len(result.Fragments) != expectedChunks {
		t.Errorf("expected %d fragments, got %d", expectedChunks, len(result.Fragments))
	}

	reassembled := e.Reassemble(result)
	if !bytes.Equal(original, reassembled) {
		t.Error("large payload reassembly failed")
	}
}

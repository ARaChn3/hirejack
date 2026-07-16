package canary

import (
	"strings"
	"sync"
	"testing"
)

func TestGenerate(t *testing.T) {
	s := NewStore()
	token := s.Generate("https://audit.example.com", "ATS-1")

	if token.ID == "" {
		t.Error("ID should not be empty")
	}
	if !strings.HasPrefix(token.ID, "hj_") {
		t.Errorf("ID should have hj_ prefix: %s", token.ID)
	}
	if token.CallbackURL == "" {
		t.Error("CallbackURL should not be empty")
	}
	if !strings.Contains(token.CallbackURL, token.ID) {
		t.Error("CallbackURL should contain token ID")
	}
	if token.Target != "ATS-1" {
		t.Errorf("Target = %q, want %q", token.Target, "ATS-1")
	}
}

func TestGenerateUnique(t *testing.T) {
	s := NewStore()

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token := s.Generate("https://example.com", "")
		if ids[token.ID] {
			t.Errorf("duplicate ID: %s", token.ID)
		}
		ids[token.ID] = true
	}
}

func TestGet(t *testing.T) {
	s := NewStore()
	token := s.Generate("https://example.com", "test")

	got := s.Get(token.ID)
	if got == nil {
		t.Fatal("Get returned nil")
	}
	if got.ID != token.ID {
		t.Error("ID mismatch")
	}
}

func TestGetNotFound(t *testing.T) {
	s := NewStore()
	got := s.Get("nonexistent")
	if got != nil {
		t.Error("should return nil for nonexistent token")
	}
}

func TestGetAll(t *testing.T) {
	s := NewStore()
	s.Generate("https://example.com", "1")
	s.Generate("https://example.com", "2")
	s.Generate("https://example.com", "3")

	all := s.GetAll()
	if len(all) != 3 {
		t.Errorf("GetAll returned %d, want 3", len(all))
	}
}

func TestRecordCallback(t *testing.T) {
	s := NewStore()
	token := s.Generate("https://example.com", "")

	cb := Callback{
		IP:        "192.168.1.1",
		UserAgent: "ATS-Parser/1.0",
	}

	ok := s.RecordCallback(token.ID, cb)
	if !ok {
		t.Error("RecordCallback should return true")
	}

	got := s.Get(token.ID)
	if len(got.Callbacks) != 1 {
		t.Errorf("should have 1 callback, got %d", len(got.Callbacks))
	}
	if got.Callbacks[0].IP != "192.168.1.1" {
		t.Error("callback IP mismatch")
	}
}

func TestRecordCallbackNotFound(t *testing.T) {
	s := NewStore()
	ok := s.RecordCallback("nonexistent", Callback{})
	if ok {
		t.Error("should return false for nonexistent token")
	}
}

func TestDelete(t *testing.T) {
	s := NewStore()
	token := s.Generate("https://example.com", "")

	ok := s.Delete(token.ID)
	if !ok {
		t.Error("Delete should return true")
	}

	got := s.Get(token.ID)
	if got != nil {
		t.Error("token should be deleted")
	}
}

func TestDeleteNotFound(t *testing.T) {
	s := NewStore()
	ok := s.Delete("nonexistent")
	if ok {
		t.Error("should return false for nonexistent token")
	}
}

func TestGenerateXXE(t *testing.T) {
	s := NewStore()
	token := s.Generate("https://audit.example.com", "")

	xxe := token.GenerateXXE()
	if !strings.Contains(xxe, "DOCTYPE") {
		t.Error("XXE should contain DOCTYPE")
	}
	if !strings.Contains(xxe, "ENTITY canary") {
		t.Error("XXE should contain ENTITY canary")
	}
	if !strings.Contains(xxe, token.CallbackURL) {
		t.Error("XXE should contain callback URL")
	}
}

func TestGenerateImageURL(t *testing.T) {
	s := NewStore()
	token := s.Generate("https://example.com", "")

	url := token.GenerateImageURL()
	if !strings.HasSuffix(url, ".png") {
		t.Error("image URL should end with .png")
	}
}

func TestHasCallbacks(t *testing.T) {
	s := NewStore()
	token := s.Generate("https://example.com", "")

	if token.HasCallbacks() {
		t.Error("should not have callbacks initially")
	}

	s.RecordCallback(token.ID, Callback{IP: "1.2.3.4"})
	token = s.Get(token.ID)

	if !token.HasCallbacks() {
		t.Error("should have callbacks after recording")
	}
}

func TestLastCallback(t *testing.T) {
	s := NewStore()
	token := s.Generate("https://example.com", "")

	if token.LastCallback() != nil {
		t.Error("LastCallback should be nil initially")
	}

	s.RecordCallback(token.ID, Callback{IP: "1.1.1.1"})
	s.RecordCallback(token.ID, Callback{IP: "2.2.2.2"})
	token = s.Get(token.ID)

	last := token.LastCallback()
	if last == nil || last.IP != "2.2.2.2" {
		t.Error("LastCallback should return most recent")
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup

	// Concurrent generates
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Generate("https://example.com", "")
		}()
	}
	wg.Wait()

	if s.Count() != 50 {
		t.Errorf("expected 50 tokens, got %d", s.Count())
	}

	// Concurrent reads and writes
	tokens := s.GetAll()
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(id string) {
			defer wg.Done()
			s.RecordCallback(id, Callback{IP: "test"})
		}(tokens[i%len(tokens)].ID)
		go func(id string) {
			defer wg.Done()
			s.Get(id)
		}(tokens[i%len(tokens)].ID)
	}
	wg.Wait()
}

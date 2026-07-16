package server

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/novusedge/hirejack/internal/pdf"
)

func setupApp() *fiber.App {
	app := fiber.New()
	srv := New("http://localhost:8080")
	srv.RegisterRoutes(app)
	return app
}

func TestGetPayloads(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/api/payloads", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("body should not be empty")
	}
}

func TestGetPayloadsByCategory(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/api/payloads/sqli", nil)
	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestGetPayloadsCategoryNotFound(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/api/payloads/nonexistent", nil)
	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestGetProfiles(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/api/profiles", nil)
	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestGeneratePersona(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/api/persona", nil)
	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var persona map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&persona)

	if persona["first_name"] == "" {
		t.Error("persona should have first_name")
	}
}

func TestInject(t *testing.T) {
	app := setupApp()

	pdfBytes := pdf.CreateTestPDF()
	pdfB64 := base64.StdEncoding.EncodeToString(pdfBytes)

	body := `{"pdf":"` + pdfB64 + `","custom_data":{"Author":"test"}}`
	req := httptest.NewRequest("POST", "/api/inject", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Errorf("status = %d, want 200, body: %s", resp.StatusCode, string(bodyBytes))
		return
	}

	var result InjectResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.PDF == "" {
		t.Error("result should have PDF")
	}
}

func TestGenerateCV(t *testing.T) {
	app := setupApp()

	body := `{"stealth":1}`
	req := httptest.NewRequest("POST", "/api/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Errorf("status = %d, body: %s", resp.StatusCode, string(bodyBytes))
		return
	}

	var result InjectResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.PDF == "" {
		t.Error("result should have PDF")
	}
}

func TestGetCallbacks(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/api/callbacks", nil)
	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestHandleCallback(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/api/c/hj_test1234", nil)
	req.Header.Set("User-Agent", "Test-Agent")

	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		t.Errorf("status = %d, want 204", resp.StatusCode)
	}
}

func TestHandleCallbackImage(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/api/c/hj_test1234.png", nil)
	resp, _ := app.Test(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", ct)
	}
}

func TestBase64Helpers(t *testing.T) {
	original := []byte("test data")

	encoded := encodeBase64(original)
	decoded, err := decodeBase64(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if string(decoded) != string(original) {
		t.Error("round-trip failed")
	}
}

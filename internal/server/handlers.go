// Package server provides HTTP API handlers for the hirejack engine.
package server

import (
	"github.com/gofiber/fiber/v2"
	"hirejack/internal/canary"
	"hirejack/internal/engine"
	"hirejack/internal/generators"
	"hirejack/internal/payloads"
	"github.com/rs/zerolog/log"
)

// Server holds the engine and provides HTTP handlers.
type Server struct {
	engine *engine.Engine
}

// New creates a new API server.
func New(canaryURL string) *Server {
	return &Server{
		engine: engine.New(canaryURL),
	}
}

// RegisterRoutes registers API routes on the Fiber app.
func (s *Server) RegisterRoutes(app *fiber.App) {
	// Docs page
	app.Get("/docs", s.ServeDocs)

	api := app.Group("/api")

	// Payloads
	api.Get("/payloads", s.GetPayloads)
	api.Get("/payloads/:category", s.GetPayloadsByCategory)
	api.Get("/profiles", s.GetProfiles)

	// Persona
	api.Get("/persona", s.GeneratePersona)

	// Inject
	api.Post("/inject", s.Inject)

	// Generate CV
	api.Post("/generate", s.GenerateCV)

	// Canary callbacks
	api.Get("/callbacks", s.GetCallbacks)
	api.Get("/c/:id.png", s.HandleCallbackImage) // More specific route first
	api.Get("/c/:id", s.HandleCallback)
}

// GetPayloads returns all payload categories.
func (s *Server) GetPayloads(c *fiber.Ctx) error {
	total := 0
	for _, cat := range payloads.Registry {
		total += len(cat.Payloads)
	}
	log.Info().
		Int("categories", len(payloads.Registry)).
		Int("payloads", total).
		Msg("payload catalog requested")
	return c.JSON(fiber.Map{"categories": payloads.Registry})
}

// GetPayloadsByCategory returns payloads for a specific category.
func (s *Server) GetPayloadsByCategory(c *fiber.Ctx) error {
	cat := payloads.GetCategory(c.Params("category"))
	if cat == nil {
		return c.Status(404).JSON(fiber.Map{"error": "category not found"})
	}
	return c.JSON(cat)
}

// GetProfiles returns available payload profiles.
func (s *Server) GetProfiles(c *fiber.Ctx) error {
	ids := payloads.ProfileIDs()
	profiles := make([]fiber.Map, len(ids))
	for i, id := range ids {
		profiles[i] = fiber.Map{
			"id":       id,
			"payloads": payloads.GetProfile(id),
		}
	}
	return c.JSON(profiles)
}

// GeneratePersona generates a random fake applicant.
func (s *Server) GeneratePersona(c *fiber.Ctx) error {
	gen := generators.NewPersonaGenerator()
	return c.JSON(gen.Generate())
}

// InjectRequest is the API request body for injection.
type InjectRequest struct {
	PDF         string            `json:"pdf"`          // Base64 encoded
	PayloadIDs  []string          `json:"payload_ids"`
	ProfileID   string            `json:"profile_id"`
	CustomData  map[string]string `json:"custom_data"`
	Stealth     int               `json:"stealth"`
	AddCanary   bool              `json:"add_canary"`
	CanaryLabel string            `json:"canary_label"`
	CanaryURL   string            `json:"canary_url"`   // Custom callback URL (for remote auditing)
}

// InjectResponse is the API response for injection.
type InjectResponse struct {
	PDF         string                  `json:"pdf"` // Base64 encoded
	Injections  []engine.InjectionPoint `json:"injections"`
	CanaryToken *canary.Token           `json:"canary_token,omitempty"`
}

// Inject applies payloads to a PDF.
func (s *Server) Inject(c *fiber.Ctx) error {
	var req InjectRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("inject: invalid request body")
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Decode base64 PDF
	pdfData, err := decodeBase64(req.PDF)
	if err != nil {
		log.Warn().Msg("inject: invalid base64 PDF")
		return c.Status(400).JSON(fiber.Map{"error": "invalid PDF encoding"})
	}

	log.Info().
		Int("pdf_size", len(pdfData)).
		Int("payloads", len(req.PayloadIDs)).
		Str("profile", req.ProfileID).
		Int("stealth", req.Stealth).
		Bool("canary", req.AddCanary).
		Msg("inject: processing")

	result, err := s.engine.Inject(engine.InjectRequest{
		PDF:         pdfData,
		PayloadIDs:  req.PayloadIDs,
		ProfileID:   req.ProfileID,
		CustomData:  req.CustomData,
		Stealth:     engine.StealthLevel(req.Stealth),
		AddCanary:   req.AddCanary,
		CanaryLabel: req.CanaryLabel,
		CanaryURL:   req.CanaryURL,
	})
	if err != nil {
		log.Error().Err(err).Msg("inject: engine failed")
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	log.Info().
		Int("injections", len(result.Injections)).
		Int("output_size", len(result.PDF)).
		Bool("has_canary", result.CanaryToken != nil).
		Msg("inject: complete")

	return c.JSON(InjectResponse{
		PDF:         encodeBase64(result.PDF),
		Injections:  result.Injections,
		CanaryToken: result.CanaryToken,
	})
}

// GenerateCVRequest is the API request for CV generation.
type GenerateCVRequest struct {
	PayloadIDs []string `json:"payload_ids"`
	ProfileID  string   `json:"profile_id"`
	Stealth    int      `json:"stealth"`
	AddCanary  bool     `json:"add_canary"`
}

// GenerateCV creates a new CV with injected payloads.
func (s *Server) GenerateCV(c *fiber.Ctx) error {
	var req GenerateCVRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	ids := req.PayloadIDs
	if req.ProfileID != "" {
		ids = payloads.GetProfile(req.ProfileID)
		log.Info().Str("profile", req.ProfileID).Int("payloads", len(ids)).Msg("generate: using profile")
	}

	log.Info().Int("payloads", len(ids)).Int("stealth", req.Stealth).Msg("generate: creating CV")

	result, err := s.engine.GenerateCV(ids, engine.StealthLevel(req.Stealth))
	if err != nil {
		log.Error().Err(err).Msg("generate: failed")
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	log.Info().Int("injections", len(result.Injections)).Int("size", len(result.PDF)).Msg("generate: complete")

	return c.JSON(InjectResponse{
		PDF:         encodeBase64(result.PDF),
		Injections:  result.Injections,
		CanaryToken: result.CanaryToken,
	})
}

// GetCallbacks returns all canary callbacks.
func (s *Server) GetCallbacks(c *fiber.Ctx) error {
	return c.JSON(s.engine.CanaryStore().GetAll())
}

// HandleCallback records a canary callback.
func (s *Server) HandleCallback(c *fiber.Ctx) error {
	id := c.Params("id")

	cb := canary.Callback{
		IP:        c.IP(),
		UserAgent: c.Get("User-Agent"),
		Headers:   make(map[string]string),
	}

	c.Request().Header.VisitAll(func(key, value []byte) {
		cb.Headers[string(key)] = string(value)
	})

	s.engine.CanaryStore().RecordCallback(id, cb)

	log.Warn().
		Str("canary_id", id).
		Str("ip", cb.IP).
		Str("user_agent", cb.UserAgent).
		Msg("🚨 canary triggered!")

	// Return 204 No Content for XXE/image callbacks
	return c.SendStatus(204)
}

// HandleCallbackImage handles image-based callbacks.
func (s *Server) HandleCallbackImage(c *fiber.Ctx) error {
	id := c.Params("id")

	cb := canary.Callback{
		IP:        c.IP(),
		UserAgent: c.Get("User-Agent"),
	}

	s.engine.CanaryStore().RecordCallback(id, cb)

	log.Warn().
		Str("canary_id", id).
		Str("ip", cb.IP).
		Str("user_agent", cb.UserAgent).
		Msg("🚨 canary image triggered!")

	// Return 1x1 transparent PNG
	c.Set("Content-Type", "image/png")
	return c.Status(200).Send(transparentPNG)
}

// 1x1 transparent PNG
var transparentPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
	0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49,
	0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
}

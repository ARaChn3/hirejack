package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/pkg/browser"
)

//go:embed frontend/dist/*
var frontend embed.FS

func main() {
	port := flag.String("port", "8080", "server port")
	headless := flag.Bool("headless", false, "don't open browser")
	flag.Parse()

	app := fiber.New(fiber.Config{
		AppName:               "hirejack",
		DisableStartupMessage: true,
	})

	app.Use(logger.New())
	app.Use(cors.New())

	// API routes
	api := app.Group("/api")
	api.Get("/payloads", handlePayloads)
	api.Post("/inject", handleInject)

	// Serve embedded frontend
	dist, err := fs.Sub(frontend, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}
	app.Use("/", filesystem.New(filesystem.Config{
		Root:         http.FS(dist),
		Browse:       false,
		Index:        "index.html",
		NotFoundFile: "index.html",
	}))

	url := fmt.Sprintf("http://localhost:%s", *port)
	log.Printf("hirejack running on %s", url)

	if !*headless {
		go browser.OpenURL(url)
	}

	log.Fatal(app.Listen(":" + *port))
}

func handlePayloads(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"categories": []fiber.Map{
			{"id": "dos", "name": "DoS", "payloads": []string{"entity_expansion", "recursion", "bloat"}},
			{"id": "sqli", "name": "SQL Injection", "payloads": []string{"basic", "blind", "union"}},
			{"id": "integrity", "name": "Data Integrity", "payloads": []string{"unicode", "truncation", "encoding"}},
			{"id": "ml_poison", "name": "ML Poisoning", "payloads": []string{"skills", "titles", "keywords"}},
		},
	})
}

func handleInject(c *fiber.Ctx) error {
	// TODO: accept PDF upload, apply selected payloads, return modified PDF
	return c.JSON(fiber.Map{"status": "not implemented"})
}

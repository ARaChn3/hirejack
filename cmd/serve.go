package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"hirejack/internal/server"
	"github.com/pkg/browser"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var frontend embed.FS

// SetFrontend sets the embedded frontend filesystem.
func SetFrontend(fs embed.FS) {
	frontend = fs
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web server",
	Long:  `Start the hirejack web server with the embedded frontend.`,
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

func init() {
	serveCmd.Flags().StringVarP(&port, "port", "p", "8080", "server port")
	serveCmd.Flags().BoolVar(&headless, "headless", false, "don't open browser")
}

func runServer() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
	})

	app := fiber.New(fiber.Config{
		AppName:               "hirejack",
		DisableStartupMessage: true,
	})

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		log.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("latency", time.Since(start)).
			Msg("")
		return err
	})
	app.Use(cors.New())

	url := fmt.Sprintf("http://localhost:%s", port)
	srv := server.New(url + "/api/c")
	srv.RegisterRoutes(app)

	dist, err := fs.Sub(frontend, "frontend/dist")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load frontend")
	}
	app.Use("/", filesystem.New(filesystem.Config{
		Root:         http.FS(dist),
		Browse:       false,
		Index:        "index.html",
		NotFoundFile: "index.html",
	}))

	log.Info().Str("url", url).Msg("server started")

	if !headless {
		go browser.OpenURL(url)
	}

	if err := app.Listen(":" + port); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}

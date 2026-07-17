package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const banner = `
 ██╗  ██╗██╗██████╗ ███████╗     ██╗ █████╗  ██████╗██╗  ██╗
 ██║  ██║██║██╔══██╗██╔════╝     ██║██╔══██╗██╔════╝██║ ██╔╝
 ███████║██║██████╔╝█████╗       ██║███████║██║     █████╔╝
 ██╔══██║██║██╔══██╗██╔══╝  ██   ██║██╔══██║██║     ██╔═██╗
 ██║  ██║██║██║  ██║███████╗╚█████╔╝██║  ██║╚██████╗██║  ██╗
 ╚═╝  ╚═╝╚═╝╚═╝  ╚═╝╚══════╝ ╚════╝ ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝
`

var (
	port     string
	headless bool
	quiet    bool
)

var rootCmd = &cobra.Command{
	Use:   "hirejack",
	Short: "ATS security auditing tool",
	Long: `hirejack is a security auditing tool for testing Applicant Tracking Systems
against various attack vectors including SQL injection, XXE, prompt injection,
and ML poisoning attacks.

It provides a web interface for uploading PDF resumes, selecting payloads,
and generating weaponized documents to test ATS parser vulnerabilities.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		if quiet {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		})
	},
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress debug output")
	rootCmd.Flags().StringVarP(&port, "port", "p", "8080", "server port")
	rootCmd.Flags().BoolVar(&headless, "headless", false, "don't open browser")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(injectCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(manCmd)
}

func Execute() {
	fmt.Print("\033[32m" + banner + "\033[0m\n")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

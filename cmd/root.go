package cmd

import (
	"fmt"
	"os"

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
)

var rootCmd = &cobra.Command{
	Use:   "hirejack",
	Short: "ATS security auditing tool",
	Long: `hirejack is a security auditing tool for testing Applicant Tracking Systems
against various attack vectors including SQL injection, XXE, prompt injection,
and ML poisoning attacks.

It provides a web interface for uploading PDF resumes, selecting payloads,
and generating weaponized documents to test ATS parser vulnerabilities.`,
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

func init() {
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

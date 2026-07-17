package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var manDir string

var manCmd = &cobra.Command{
	Use:    "man",
	Short:  "Generate man pages",
	Long:   `Generate man pages for hirejack and all subcommands.`,
	Hidden: true,
	Run:    runMan,
}

func init() {
	manCmd.Flags().StringVarP(&manDir, "dir", "d", "./man", "output directory")
}

func runMan(cmd *cobra.Command, args []string) {
	if err := os.MkdirAll(manDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}

	header := &doc.GenManHeader{
		Title:   "HIREJACK",
		Section: "1",
		Source:  "hirejack",
		Manual:  "User Commands",
	}

	if err := doc.GenManTree(rootCmd, header, manDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating man pages: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated man pages in %s/\n", manDir)
}

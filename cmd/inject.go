package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"hirejack/internal/engine"
	"hirejack/internal/payloads"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	outputPath  string
	payloadIDs  []string
	profileID   string
	stealthLvl  int
	addCanary   bool
	canaryLabel string
)

var injectCmd = &cobra.Command{
	Use:   "inject [pdf-file]",
	Short: "Inject payloads into a PDF",
	Long: `Inject selected payloads into a PDF file from the command line.

Examples:
  hirejack inject resume.pdf -o poisoned.pdf --payloads sqli.name_inject,xxe.dtd_external
  hirejack inject resume.pdf --profile full_audit --stealth 3
  hirejack inject resume.pdf --payloads llm_injection.ignore_resume --canary`,
	Args: cobra.ExactArgs(1),
	Run:  runInject,
}

func init() {
	injectCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output path (default: input__hirejacked.pdf)")
	injectCmd.Flags().StringSliceVarP(&payloadIDs, "payloads", "P", nil, "comma-separated payload IDs")
	injectCmd.Flags().StringVar(&profileID, "profile", "", "use payload profile (full_audit, sqli_focus, etc)")
	injectCmd.Flags().IntVarP(&stealthLvl, "stealth", "s", 1, "stealth level 1-5")
	injectCmd.Flags().BoolVar(&addCanary, "canary", false, "add canary token")
	injectCmd.Flags().StringVar(&canaryLabel, "canary-label", "", "label for canary token")
}

func runInject(cmd *cobra.Command, args []string) {
	inputPath := args[0]

	pdfData, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatal().Err(err).Str("file", inputPath).Msg("failed to read input")
	}

	if outputPath == "" {
		ext := filepath.Ext(inputPath)
		base := strings.TrimSuffix(inputPath, ext)
		outputPath = base + "__hirejacked.pdf"
	}

	var ids []string
	if profileID != "" {
		ids = payloads.GetProfile(profileID)
		if ids == nil {
			log.Fatal().Str("profile", profileID).Msg("unknown profile")
		}
		fmt.Printf("Using profile: %s (%d payloads)\n", profileID, len(ids))
	}
	ids = append(ids, payloadIDs...)

	if len(ids) == 0 {
		log.Fatal().Msg("no payloads specified (use --payloads or --profile)")
	}

	fmt.Printf("Input:    %s (%d bytes)\n", inputPath, len(pdfData))
	fmt.Printf("Payloads: %d\n", len(ids))
	fmt.Printf("Stealth:  L%d\n", stealthLvl)

	eng := engine.New("http://localhost:8080/api/c")
	result, err := eng.Inject(engine.InjectRequest{
		PDF:         pdfData,
		PayloadIDs:  ids,
		Stealth:     engine.StealthLevel(stealthLvl),
		AddCanary:   addCanary,
		CanaryLabel: canaryLabel,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("injection failed")
	}

	if err := os.WriteFile(outputPath, result.PDF, 0644); err != nil {
		log.Fatal().Err(err).Str("file", outputPath).Msg("failed to write output")
	}

	fmt.Printf("Output:   %s (%d bytes)\n", outputPath, len(result.PDF))
	fmt.Printf("Injected: %d points\n", len(result.Injections))
	if result.CanaryToken != nil {
		fmt.Printf("Canary:   %s\n", result.CanaryToken.CallbackURL)
	}
}

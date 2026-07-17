package cmd

import (
	"fmt"

	"hirejack/internal/payloads"
	"github.com/spf13/cobra"
)

var (
	showCategory string
	showProfiles bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available payloads",
	Long: `List all available payload categories and their payloads.

Examples:
  hirejack list                    List all categories with payload counts
  hirejack list -c sqli            List payloads in sqli category
  hirejack list --profiles         List available profiles`,
	Run: runList,
}

func init() {
	listCmd.Flags().StringVarP(&showCategory, "category", "c", "", "show payloads in category")
	listCmd.Flags().BoolVar(&showProfiles, "profiles", false, "show available profiles")
}

func runList(cmd *cobra.Command, args []string) {
	if showProfiles {
		fmt.Println("\033[1mPROFILES\033[0m")
		for _, id := range payloads.ProfileIDs() {
			pids := payloads.GetProfile(id)
			fmt.Printf("  %-20s %d payloads\n", id, len(pids))
		}
		return
	}

	if showCategory != "" {
		cat := payloads.GetCategory(showCategory)
		if cat == nil {
			fmt.Printf("Unknown category: %s\n", showCategory)
			return
		}
		fmt.Printf("\033[1m%s\033[0m (%s)\n", cat.Name, cat.ID)
		fmt.Printf("%s\n\n", cat.Description)
		for _, p := range cat.Payloads {
			impact := colorImpact(p.Impact)
			stealth := formatStealth(p.Stealth)
			fmt.Printf("  \033[36m%-25s\033[0m %s %s\n", p.ID, impact, stealth)
			if p.Description != "" {
				fmt.Printf("    %s\n", p.Description)
			}
		}
		return
	}

	fmt.Println("\033[1mPAYLOAD CATEGORIES\033[0m")
	total := 0
	for _, cat := range payloads.Registry {
		fmt.Printf("  \033[33m%-15s\033[0m %3d payloads   %s\n", cat.ID, len(cat.Payloads), cat.Name)
		total += len(cat.Payloads)
	}
	fmt.Printf("\n  Total: %d payloads\n", total)
	fmt.Println("\n  Use -c <category> to see payloads in a category")
}

func colorImpact(impact string) string {
	switch impact {
	case "critical":
		return "\033[31m[CRIT]\033[0m"
	case "high":
		return "\033[33m[HIGH]\033[0m"
	case "medium":
		return "\033[34m[MED] \033[0m"
	default:
		return "\033[90m[LOW] \033[0m"
	}
}

func formatStealth(levels []int) string {
	if len(levels) == 0 {
		return ""
	}
	min, max := levels[0], levels[len(levels)-1]
	if min == max {
		return fmt.Sprintf("L%d", min)
	}
	return fmt.Sprintf("L%d-%d", min, max)
}

package config

import (
	"fmt"
	"strings"

	"github.com/muesli/termenv"
)

var output = termenv.ColorProfile()

func green(s string) string  { return termenv.String(s).Foreground(output.Color("2")).String() }
func yellow(s string) string { return termenv.String(s).Foreground(output.Color("3")).String() }
func bold(s string) string   { return termenv.String(s).Bold().String() }
func dim(s string) string    { return termenv.String(s).Faint().String() }

// printAudit prints the full security audit report.
func printAudit(result AuditResult) {
	fmt.Println(bold("Security Audit - Effective Configuration"))
	fmt.Println(bold("========================================="))
	fmt.Println()

	for _, gr := range result.Groups {
		printGroupHeader(gr)
		printGroupKeys(gr.Keys)
		fmt.Println()
	}

	printSummaryLine(result)
}

// printGroupHeader prints the group name with posture icon and tags.
func printGroupHeader(gr AuditGroupResult) {
	var icon string
	var tagStr string
	tags := strings.Join(gr.Posture.Tags, " ")

	if gr.Posture.Secure {
		icon = green("✓")
		tagStr = green(tags)
	} else {
		icon = yellow("⚠")
		tagStr = yellow(tags)
	}

	fmt.Printf("%-16s %s %s\n", bold(gr.Group.Name), icon, tagStr)
}

// printGroupKeys prints the resolved keys for a group with aligned columns.
func printGroupKeys(keys []ResolvedKey) {
	for _, rk := range keys {
		value := rk.Value
		if value == "" {
			value = "-"
		}
		source := rk.Source
		if source == "" {
			source = "-"
		}

		sourceStr := dim("(" + source + ")")
		if source == "project" || source == "global" || source == "env" {
			sourceStr = bold("(" + source + ")")
		}

		fmt.Printf("  %-28s %-16s %s\n", dim(rk.Key), value, sourceStr)
	}
}

// printSummaryLine prints the overall audit summary.
func printSummaryLine(result AuditResult) {
	if result.RelaxedCount == 0 {
		fmt.Printf("%s %s\n", bold("Overall:"),
			green(fmt.Sprintf("✓ All %d categories are locked down", result.TotalGroups)))
	} else {
		fmt.Printf("%s %s\n", bold("Overall:"),
			yellow(fmt.Sprintf("⚠ %d of %d categories have relaxed settings",
				result.RelaxedCount, result.TotalGroups)))
	}
}

package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

// ConsoleFormatter outputs scan results as colorized CLI output.
//
// This is the default format for interactive terminal use, providing
// human-readable output with color coding for drift severity.
type ConsoleFormatter struct{}

// NewConsoleFormatter creates a new console formatter.
func NewConsoleFormatter() *ConsoleFormatter {
	return &ConsoleFormatter{}
}

// Format writes the scan result as colorized text to the provided writer.
func (f *ConsoleFormatter) Format(w io.Writer, result ScanResult) error {
	// Summary header
	if result.DriftCount == 0 {
		fmt.Fprintf(w, "\n%s\n", color.GreenString("✅ No drift detected!"))
		fmt.Fprintf(w, "   Scanned %d %s resources in %dms\n\n",
			result.TotalResources, result.Service, result.ScanDuration)
		return nil
	}

	fmt.Fprintf(w, "\n%s\n", color.YellowString("⚠️  Drift detected!"))
	fmt.Fprintf(w, "   %d of %d %s resources have drift\n\n",
		result.DriftCount, result.TotalResources, result.Service)

	// Print each drift
	for i, drift := range result.Drifts {
		if !drift.HasDrift() {
			continue
		}

		// Resource header
		fmt.Fprintf(w, "%s\n", color.CyanString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Fprintf(w, "📦 %s\n", color.WhiteString("%s (%s)", drift.ResourceName, drift.ResourceType))

		if drift.Missing {
			fmt.Fprintf(w, "   %s\n", color.RedString("❌ MISSING - Resource not found in AWS"))
			continue
		}

		// Attribute diffs
		if len(drift.Diffs) > 0 {
			fmt.Fprintf(w, "   %s\n", color.YellowString("Attribute differences:"))
			for attr, values := range drift.Diffs {
				expected := f.formatValue(values[0])
				actual := f.formatValue(values[1])
				fmt.Fprintf(w, "     • %s:\n", color.WhiteString(attr))
				fmt.Fprintf(w, "       %s %s\n", color.RedString("- expected:"), expected)
				fmt.Fprintf(w, "       %s %s\n", color.GreenString("+ actual:  "), actual)
			}
		}

		// Extra attributes
		if len(drift.ExtraAttributes) > 0 {
			fmt.Fprintf(w, "   %s\n", color.BlueString("Extra attributes in AWS:"))
			for attr, value := range drift.ExtraAttributes {
				fmt.Fprintf(w, "     • %s: %s\n", color.WhiteString(attr), f.formatValue(value))
			}
		}

		if i < len(result.Drifts)-1 {
			fmt.Fprintln(w)
		}
	}

	fmt.Fprintf(w, "%s\n\n", color.CyanString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))

	// Summary footer
	fmt.Fprintf(w, "📊 Summary: %s resources with drift out of %d scanned\n",
		color.YellowString("%d", result.DriftCount), result.TotalResources)
	fmt.Fprintf(w, "⏱️  Scan completed in %dms\n\n", result.ScanDuration)

	return nil
}

func (f *ConsoleFormatter) formatValue(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return color.HiBlackString("<not set>")
	case string:
		if val == "" {
			return color.HiBlackString("<empty>")
		}
		return fmt.Sprintf("%q", val)
	case bool:
		if val {
			return color.GreenString("true")
		}
		return color.RedString("false")
	case map[string]interface{}:
		if len(val) == 0 {
			return color.HiBlackString("{}")
		}
		parts := make([]string, 0, len(val))
		for k, v := range val {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
		return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
	case []interface{}:
		if len(val) == 0 {
			return color.HiBlackString("[]")
		}
		parts := make([]string, len(val))
		for i, v := range val {
			parts[i] = fmt.Sprintf("%v", v)
		}
		return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Name returns the format name.
func (f *ConsoleFormatter) Name() string {
	return "console"
}

// FileExtension returns the recommended file extension.
func (f *ConsoleFormatter) FileExtension() string {
	return ".txt"
}

func init() {
	Register(FormatConsole, NewConsoleFormatter())
}

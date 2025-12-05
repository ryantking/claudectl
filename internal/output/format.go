package output

import (
	"encoding/json"
	"fmt"
	"os"
)

var jsonOutput bool

// SetJSONOutput sets whether output should be in JSON format.
func SetJSONOutput(enabled bool) {
	jsonOutput = enabled
}

// IsJSONOutput returns whether JSON output is enabled.
func IsJSONOutput() bool {
	return jsonOutput
}

// FormatJSON formats a result as JSON.
func FormatJSON(result Result) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode JSON: %v\n", err)
	}
}

// FormatHuman formats a result for human-readable output.
func FormatHuman(result Result) {
	if result.Success {
		if result.Message != "" {
			fmt.Println(result.Message)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Message)
	}
}

// Output formats and outputs a result based on the current output mode.
func Output(result Result) {
	if jsonOutput {
		FormatJSON(result)
	} else {
		FormatHuman(result)
	}
}

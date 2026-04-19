package helmstrip

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// yamlBooleans are values YAML 1.1 and 1.2 interpret as booleans.
var yamlBooleans = map[string]bool{
	"true": true, "false": true,
	"True": true, "False": true,
	"TRUE": true, "FALSE": true,
	"yes": true, "no": true,
	"Yes": true, "No": true,
	"YES": true, "NO": true,
	"on": true, "off": true,
	"On": true, "Off": true,
	"ON": true, "OFF": true,
	"y": true, "n": true,
	"Y": true, "N": true,
}

// yamlNulls are values YAML interprets as null.
var yamlNulls = map[string]bool{
	"null": true, "Null": true, "NULL": true,
	"~": true, "": true,
}

// specialLeadingChars are characters that, when appearing at the start of a
// plain scalar, would cause YAML parsing issues.
const specialLeadingChars = "-:#{[]*&!%@?>`|,'\""

// reInteger matches YAML integer forms: decimal, hex, octal, binary.
var reInteger = regexp.MustCompile(`^[-+]?(0|[1-9][0-9]*|0x[0-9a-fA-F]+|0o[0-7]+|0b[01]+)$`)

// reFloat matches YAML float forms including scientific notation.
var reFloat = regexp.MustCompile(`^[-+]?(\.[0-9]+|[0-9]+(\.[0-9]*)?)([eE][-+]?[0-9]+)?$`)

// reSpecialFloat matches YAML special float values.
var reSpecialFloat = regexp.MustCompile(`(?i)^[-+]?\.(inf|nan)$`)

// reTimestamp loosely matches date/datetime patterns that YAML may interpret.
var reTimestamp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`)

// UnquoteScalars walks a yaml.Node tree and switches quoted scalar nodes to
// plain style when the value is unambiguous without quotes.
func UnquoteScalars(node *yaml.Node) {
	walkNodes(node, func(n *yaml.Node) {
		if n.Kind != yaml.ScalarNode {
			return
		}
		// Only process quoted scalars (single or double).
		if n.Style != yaml.SingleQuotedStyle && n.Style != yaml.DoubleQuotedStyle {
			return
		}
		if needsQuoting(n.Value) {
			return
		}
		n.Style = 0 // plain style
	})
}

// needsQuoting returns true if the value would be ambiguous or broken as a
// plain YAML scalar.
func needsQuoting(s string) bool {
	if s == "" {
		return true
	}

	if yamlBooleans[s] || yamlNulls[s] {
		return true
	}

	if reInteger.MatchString(s) || reFloat.MatchString(s) || reSpecialFloat.MatchString(s) {
		return true
	}

	if reTimestamp.MatchString(s) {
		return true
	}

	// Check leading character.
	if strings.ContainsRune(specialLeadingChars, rune(s[0])) {
		return true
	}

	// Check for mid-string sequences that break plain scalars.
	if strings.Contains(s, ": ") || strings.Contains(s, " #") || strings.Contains(s, ", ") {
		return true
	}

	// Strings with newlines, tabs, or control characters.
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' || r < 0x20 {
			return true
		}
	}

	// Leading or trailing whitespace.
	if s[0] == ' ' || s[len(s)-1] == ' ' {
		return true
	}

	// Multiple consecutive spaces (could be significant).
	if strings.Contains(s, "  ") {
		return true
	}

	return false
}

// walkNodes recursively visits all nodes in a yaml.Node tree.
func walkNodes(node *yaml.Node, fn func(*yaml.Node)) {
	if node == nil {
		return
	}
	fn(node)
	for _, child := range node.Content {
		walkNodes(child, fn)
	}
}

package badge

import (
	"fmt"
	"net/url"
	"strings"
)

// namedColors maps Shields.io-compatible named colors to hex values (without #).
var namedColors = map[string]string{
	"brightgreen":   "4c1",
	"green":         "97ca00",
	"yellowgreen":   "a4a61d",
	"yellow":        "dfb317",
	"orange":        "fe7d37",
	"red":           "e05d44",
	"blue":          "007ec6",
	"lightgrey":     "9f9f9f",
	"lightgray":     "9f9f9f",
	"grey":          "9f9f9f",
	"gray":          "9f9f9f",
	"success":       "4c1",
	"important":     "fe7d37",
	"critical":      "e05d44",
	"informational": "007ec6",
	"inactive":      "9f9f9f",
}

// decodeBadgeField applies Shields.io text-encoding rules to a single path field:
//   - "__" → literal "_"
//   - "_"  → space
func decodeBadgeField(s string) string {
	const literalUnderscore = "\x00"
	s = strings.ReplaceAll(s, "__", literalUnderscore)
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, literalUnderscore, "_")
	return s
}

// ResolveColor returns the 3- or 6-digit hex string (no leading #) for a named
// color or a bare hex value. It returns an error if the input is not recognised.
func ResolveColor(color string) (string, error) {
	if color == "" {
		return "", fmt.Errorf("color must not be empty")
	}
	if hex, ok := namedColors[strings.ToLower(color)]; ok {
		return hex, nil
	}
	// Accept hex with or without leading '#'
	c := strings.TrimPrefix(color, "#")
	if len(c) == 3 || len(c) == 6 {
		valid := true
		for _, ch := range c {
			if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
				valid = false
				break
			}
		}
		if valid {
			return strings.ToLower(c), nil
		}
	}
	return "", fmt.Errorf("invalid color %q: use a named color (e.g. brightgreen) or hex value (e.g. 007ec6)", color)
}

// ParsePath parses the badge path segment that follows "/badge/".
//
// It supports two formats:
//   - label-message-color  (three or more dash-separated fields)
//   - message-color        (exactly two dash-separated fields)
//
// Shields.io path-encoding rules apply:
//   - "--" encodes a literal "-" in the text
//   - "_"  encodes a space; "__" encodes a literal "_"
//
// URL percent-encoding (e.g. %20) is decoded before any further processing.
//
// The returned color is a resolved hex string (no leading #).
func ParsePath(rawPath string) (label, message, color string, err error) {
	if rawPath == "" {
		return "", "", "", fmt.Errorf("badge path must not be empty")
	}

	// Decode percent-encoding first (handles %20 etc.)
	decoded, decodeErr := url.PathUnescape(rawPath)
	if decodeErr != nil {
		decoded = rawPath
	}

	// Preserve double-hyphens so they survive the single-hyphen split.
	// Use a two-rune placeholder that cannot appear in valid badge text
	// (both runes are Unicode non-characters and cannot be encoded in URLs).
	const hyphenPlaceholder = "\uFFFE\uFFFF"
	s := strings.ReplaceAll(decoded, "--", hyphenPlaceholder)
	parts := strings.Split(s, "-")

	// Restore double-hyphens (now a literal "-" within each field)
	for i, p := range parts {
		parts[i] = strings.ReplaceAll(p, hyphenPlaceholder, "-")
	}

	switch len(parts) {
	case 1:
		return "", "", "", fmt.Errorf("badge path must contain at least one hyphen separating message and color")
	case 2:
		message = decodeBadgeField(parts[0])
		color = parts[1]
	default:
		// Three or more parts: last = color, second-to-last = message, rest = label
		color = parts[len(parts)-1]
		message = decodeBadgeField(parts[len(parts)-2])
		labelParts := make([]string, len(parts)-2)
		for i, p := range parts[:len(parts)-2] {
			labelParts[i] = decodeBadgeField(p)
		}
		label = strings.Join(labelParts, "-")
	}

	if message == "" {
		return "", "", "", fmt.Errorf("badge message must not be empty")
	}

	resolvedColor, colorErr := ResolveColor(color)
	if colorErr != nil {
		return "", "", "", colorErr
	}

	return label, message, resolvedColor, nil
}

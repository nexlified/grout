package render

import (
	"bytes"
	"fmt"
	"strings"
)

// BadgeStyle controls the visual style of a generated badge.
type BadgeStyle string

const (
	// BadgeStyleFlat is the default flat style with rounded corners.
	BadgeStyleFlat BadgeStyle = "flat"
	// BadgeStyleFlatSquare is the flat style without rounded corners.
	BadgeStyleFlatSquare BadgeStyle = "flat-square"
	// BadgeStyleForTheBadge is a larger, all-uppercase style.
	BadgeStyleForTheBadge BadgeStyle = "for-the-badge"
)

// badgeDimensions returns (fontSize, badgeHeight, hPadding) for a given style.
func badgeDimensions(style BadgeStyle) (fontSize float64, height int, hPadding float64) {
	switch style {
	case BadgeStyleForTheBadge:
		return 10, 28, 12
	default:
		return 11, 20, 6
	}
}

// estimateBadgeTextWidth returns a rough pixel width for a string rendered at
// the given font size. The SVG uses "DejaVu Sans,Verdana,Geneva,sans-serif",
// and 0.65 × fontSize per character is a reasonable approximation for that
// stack; actual width may vary slightly depending on the client's available fonts.
func estimateBadgeTextWidth(text string, fontSize float64) float64 {
	return float64(len(text)) * fontSize * 0.65
}

// DrawBadge generates a Shields-style flat SVG badge.
//
//   - label        – left-side text (may be empty for a message-only badge)
//   - message      – right-side text (required)
//   - labelColorHex  – 3- or 6-digit hex for the label background (no "#"); defaults to "555"
//   - messageColorHex – 3- or 6-digit hex for the message background (no "#"); defaults to "007ec6"
//   - style        – visual style (flat, flat-square, for-the-badge)
func (r *Renderer) DrawBadge(label, message, labelColorHex, messageColorHex string, style BadgeStyle) ([]byte, error) {
	fontSize, height, hPadding := badgeDimensions(style)

	// For-the-badge style uses uppercase text
	if style == BadgeStyleForTheBadge {
		label = strings.ToUpper(label)
		message = strings.ToUpper(message)
	}

	// Default colors
	if labelColorHex == "" {
		labelColorHex = "555"
	}
	if messageColorHex == "" {
		messageColorHex = "007ec6"
	}

	hasLabel := label != ""

	// Calculate section widths
	const minSectionWidth = 20.0
	labelW := 0.0
	if hasLabel {
		labelW = estimateBadgeTextWidth(label, fontSize) + hPadding*2
		if labelW < minSectionWidth {
			labelW = minSectionWidth
		}
	}
	messageW := estimateBadgeTextWidth(message, fontSize) + hPadding*2
	if messageW < minSectionWidth {
		messageW = minSectionWidth
	}

	totalW := labelW + messageW

	// Rounded corners for flat style; square for flat-square and for-the-badge
	rx := 3
	if style == BadgeStyleFlatSquare || style == BadgeStyleForTheBadge {
		rx = 0
	}

	// SVG text baseline: sits at ~70% of badge height
	textY := int(float64(height) * 0.70)
	shadowY := textY + 1

	labelCX := int(labelW / 2)
	messageCX := int(labelW + messageW/2)

	// Accessible label for the aria-label attribute
	ariaLabel := message
	if hasLabel {
		ariaLabel = label + ": " + message
	}

	var buf bytes.Buffer
	tw := int(totalW)

	buf.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" role="img" aria-label="%s">`,
		tw, height, escapeXML(ariaLabel)))
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf(`<title>%s</title>`, escapeXML(ariaLabel)))
	buf.WriteString("\n")

	// Gradient overlay for flat style only
	if rx > 0 {
		buf.WriteString(`<linearGradient id="s" x2="0" y2="100%">`)
		buf.WriteString(`<stop offset="0" stop-color="#bbb" stop-opacity=".1"/>`)
		buf.WriteString(`<stop offset="1" stop-opacity=".1"/>`)
		buf.WriteString(`</linearGradient>`)
		buf.WriteString("\n")
		buf.WriteString(fmt.Sprintf(`<clipPath id="r"><rect width="%d" height="%d" rx="%d" fill="#fff"/></clipPath>`,
			tw, height, rx))
		buf.WriteString("\n")
	}

	// Background rectangles
	if rx > 0 {
		buf.WriteString(`<g clip-path="url(#r)">`)
	} else {
		buf.WriteString(`<g>`)
	}
	buf.WriteString("\n")

	if hasLabel {
		buf.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#%s"/>`,
			int(labelW), height, labelColorHex))
		buf.WriteString("\n")
	}
	buf.WriteString(fmt.Sprintf(`<rect x="%d" width="%d" height="%d" fill="#%s"/>`,
		int(labelW), int(messageW), height, messageColorHex))
	buf.WriteString("\n")

	if rx > 0 {
		buf.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="url(#s)"/>`, tw, height))
		buf.WriteString("\n")
	}
	buf.WriteString(`</g>`)
	buf.WriteString("\n")

	// Text
	buf.WriteString(fmt.Sprintf(
		`<g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="%.0f">`,
		fontSize,
	))
	buf.WriteString("\n")

	if hasLabel {
		buf.WriteString(fmt.Sprintf(
			`<text x="%d" y="%d" fill="#010101" fill-opacity=".3" aria-hidden="true">%s</text>`,
			labelCX, shadowY, escapeXML(label)))
		buf.WriteString("\n")
		buf.WriteString(fmt.Sprintf(
			`<text x="%d" y="%d">%s</text>`,
			labelCX, textY, escapeXML(label)))
		buf.WriteString("\n")
	}

	buf.WriteString(fmt.Sprintf(
		`<text x="%d" y="%d" fill="#010101" fill-opacity=".3" aria-hidden="true">%s</text>`,
		messageCX, shadowY, escapeXML(message)))
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf(
		`<text x="%d" y="%d">%s</text>`,
		messageCX, textY, escapeXML(message)))
	buf.WriteString("\n")

	buf.WriteString(`</g>`)
	buf.WriteString("\n")
	buf.WriteString(`</svg>`)

	return buf.Bytes(), nil
}

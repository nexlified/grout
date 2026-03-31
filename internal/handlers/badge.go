package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"grout/internal/badge"
	"grout/internal/render"
)

// handleBadge serves Shields-style SVG badge images from URLs such as:
//
//	/badge/build-passing-brightgreen
//	/badge/just%20the%20message-8A2BE2
//
// Supported query parameters:
//   - label       – override the label (left-side) text
//   - labelColor  – label background color (named or hex, default #555)
//   - color       – message background color override (named or hex)
//   - style       – flat (default), flat-square, for-the-badge
//   - cacheSeconds – accepted but does not affect server-side cache headers
//   - logo, logoColor, link – accepted and silently ignored
func (s *Service) handleBadge(w http.ResponseWriter, r *http.Request) {
	pathContent := strings.TrimPrefix(r.URL.Path, "/badge/")
	if pathContent == "" {
		s.serveErrorPage(w, http.StatusBadRequest,
			"Badge path is required. Use /badge/message-color or /badge/label-message-color.")
		return
	}

	// Parse the badge path
	label, message, colorHex, err := badge.ParsePath(pathContent)
	if err != nil {
		s.serveErrorPage(w, http.StatusBadRequest, fmt.Sprintf("Invalid badge URL: %s", err))
		return
	}

	// Query parameter overrides
	if qLabel := r.URL.Query().Get("label"); qLabel != "" {
		label = qLabel
	}
	if qColor := r.URL.Query().Get("color"); qColor != "" {
		if resolved, resolveErr := badge.ResolveColor(qColor); resolveErr == nil {
			colorHex = resolved
		}
	}

	labelColorHex := "555"
	if qLabelColor := r.URL.Query().Get("labelColor"); qLabelColor != "" {
		if resolved, resolveErr := badge.ResolveColor(qLabelColor); resolveErr == nil {
			labelColorHex = resolved
		}
	}

	styleParam := r.URL.Query().Get("style")
	style := render.BadgeStyle(styleParam)
	switch style {
	case render.BadgeStyleFlat, render.BadgeStyleFlatSquare, render.BadgeStyleForTheBadge:
		// valid
	default:
		style = render.BadgeStyleFlat
	}

	cacheKey := fmt.Sprintf("Badge:%s:%s:%s:%s:%s", label, message, colorHex, labelColorHex, style)
	s.serveImage(w, r, cacheKey, render.FormatSVG, func() ([]byte, error) {
		return s.renderer.DrawBadge(label, message, labelColorHex, colorHex, style)
	})
}

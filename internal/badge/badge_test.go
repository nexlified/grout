package badge

import (
	"testing"
)

func TestParsePath(t *testing.T) {
	cases := []struct {
		name        string
		input       string
		wantLabel   string
		wantMessage string
		wantColor   string
		wantErr     bool
	}{
		// Basic two-field format: message-color
		{
			name:        "message and named color",
			input:       "passing-brightgreen",
			wantMessage: "passing",
			wantColor:   "4c1",
		},
		{
			name:        "message and hex color",
			input:       "passing-007ec6",
			wantMessage: "passing",
			wantColor:   "007ec6",
		},
		{
			name:        "message and 3-digit hex color",
			input:       "passing-4c1",
			wantMessage: "passing",
			wantColor:   "4c1",
		},
		{
			name:        "message with percent-encoded space",
			input:       "just%20the%20message-8a2be2",
			wantMessage: "just the message",
			wantColor:   "8a2be2",
		},
		{
			name:        "underscore encodes space",
			input:       "hello_world-red",
			wantMessage: "hello world",
			wantColor:   "e05d44",
		},
		{
			name:        "double-underscore is literal underscore",
			input:       "hello__world-blue",
			wantMessage: "hello_world",
			wantColor:   "007ec6",
		},
		// Three-field format: label-message-color
		{
			name:        "label message color",
			input:       "build-passing-brightgreen",
			wantLabel:   "build",
			wantMessage: "passing",
			wantColor:   "4c1",
		},
		{
			name:        "label message hex color",
			input:       "build-passing-4c1",
			wantLabel:   "build",
			wantMessage: "passing",
			wantColor:   "4c1",
		},
		// Double-hyphen encodes literal hyphen in label
		{
			name:        "double-hyphen in label becomes hyphen",
			input:       "my--app-passing-green",
			wantLabel:   "my-app",
			wantMessage: "passing",
			wantColor:   "97ca00",
		},
		// Named colors
		{
			name:        "named color success",
			input:       "tests-passing-success",
			wantLabel:   "tests",
			wantMessage: "passing",
			wantColor:   "4c1",
		},
		{
			name:        "named color critical",
			input:       "build-failing-critical",
			wantLabel:   "build",
			wantMessage: "failing",
			wantColor:   "e05d44",
		},
		{
			name:        "named color lightgrey alias",
			input:       "status-unknown-lightgray",
			wantLabel:   "status",
			wantMessage: "unknown",
			wantColor:   "9f9f9f",
		},
		// Case insensitivity for named colors
		{
			name:        "named color case insensitive",
			input:       "build-ok-BrightGreen",
			wantLabel:   "build",
			wantMessage: "ok",
			wantColor:   "4c1",
		},
		// Hex color with # prefix (should be accepted)
		{
			name:        "hex color with hash prefix",
			input:       "v1-stable-%238A2BE2",
			wantLabel:   "v1",
			wantMessage: "stable",
			wantColor:   "8a2be2",
		},
		// Error cases
		{
			name:    "empty path",
			input:   "",
			wantErr: true,
		},
		{
			name:    "no hyphen",
			input:   "nodash",
			wantErr: true,
		},
		{
			name:    "invalid color",
			input:   "build-passing-notacolor",
			wantErr: true,
		},
		{
			name:    "empty message",
			input:   "--red",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			label, message, color, err := ParsePath(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error but got label=%q message=%q color=%q", label, message, color)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if label != tc.wantLabel {
				t.Errorf("label: want %q, got %q", tc.wantLabel, label)
			}
			if message != tc.wantMessage {
				t.Errorf("message: want %q, got %q", tc.wantMessage, message)
			}
			if color != tc.wantColor {
				t.Errorf("color: want %q, got %q", tc.wantColor, color)
			}
		})
	}
}

func TestResolveColor(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantHex string
		wantErr bool
	}{
		{"brightgreen named", "brightgreen", "4c1", false},
		{"red named", "red", "e05d44", false},
		{"blue named", "blue", "007ec6", false},
		{"lightgrey alias", "lightgrey", "9f9f9f", false},
		{"lightgray alias", "lightgray", "9f9f9f", false},
		{"grey alias", "grey", "9f9f9f", false},
		{"success alias", "success", "4c1", false},
		{"6-digit hex", "ff5733", "ff5733", false},
		{"3-digit hex", "f53", "f53", false},
		{"6-digit hex uppercase", "FF5733", "ff5733", false},
		{"hex with hash", "#ff5733", "ff5733", false},
		{"empty string", "", "", true},
		{"invalid hex", "gggggg", "", true},
		{"invalid name", "neon", "", true},
		{"5-digit hex", "ff573", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveColor(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error but got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantHex {
				t.Errorf("want %q, got %q", tc.wantHex, got)
			}
		})
	}
}

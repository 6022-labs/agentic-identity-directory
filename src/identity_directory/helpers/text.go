package helpers

import "strings"

// SanitizeText drops NUL and invalid UTF-8, which Postgres refuses: one such on-chain string would stall a chain.
func SanitizeText(text string) string {
	return strings.ReplaceAll(strings.ToValidUTF8(text, "�"), "\x00", "")
}

func SanitizeTexts(texts map[string]string) map[string]string {
	sanitized := make(map[string]string, len(texts))
	for key, value := range texts {
		sanitized[SanitizeText(key)] = SanitizeText(value)
	}
	return sanitized
}

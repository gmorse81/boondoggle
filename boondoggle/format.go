package boondoggle

import (
	"regexp"
	"strings"
)

var (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"

	pwRegex    *regexp.Regexp
	emailRegex *regexp.Regexp
)

func init() {
	pwRegex = regexp.MustCompile(`(?:password=)([^\s]+)`)
	emailRegex = regexp.MustCompile(`(?:email=)([^\s]+)`)
}

// Format wraps a given message in a given color,
// and obfuscates some sensitive output.
func Format(color, message string) string {
	message = obfuscate(message)
	return color + message + Reset
}

func obfuscate(message string) string {
	// Run regexes against the message.
	message = pwRegex.ReplaceAllStringFunc(message, func(m string) string {
		src := strings.Split(m, "=")
		m = strings.ReplaceAll(m, src[1], "******")
		return m
	})
	message = emailRegex.ReplaceAllStringFunc(message, func(m string) string {
		src := strings.Split(m, "=")
		m = strings.ReplaceAll(m, src[1], "******")
		return m
	})
	return message
}

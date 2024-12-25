package cleaner

import (
  "regexp"
  "strings"
)

func SanitizeLogMessage(message string) string {
  // Remove non-printable characters using a regex
  re := regexp.MustCompile(`[[:cntrl:]]`)
  clean := strings.TrimSpace(re.ReplaceAllString(message, ""))
  return StripAnsiCodes(clean)
}

func StripAnsiCodes(input string) string {
  // Strip ANSI codes from a string
  var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
  return ansiRegex.ReplaceAllString(input, "")
}

//go:build !dot

package display

import (
	"strings"
	"unicode"
)

// wrapText preserves paragraph breaks and prefers spaces between words. A word
// that cannot fit on an empty line (including Japanese text without spaces) is
// split by measured width, without splitting UTF-8 or separating combining marks.
// A single character wider than the available space is emitted alone.
func wrapText(text string, width int, measure func(string) int) []string {
	if text == "" {
		return nil
	}
	var lines []string
	for _, paragraph := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		line := ""
		for _, word := range words {
			if line != "" {
				if measure(line+" "+word) <= width {
					line += " " + word
					continue
				}
				lines = append(lines, line)
				line = ""
			}
			for {
				cut := 0
				overflow := false
				var previous rune
				for end, next := range word {
					joined := previous == '\u200d'
					previous = next
					if end == 0 || unicode.IsMark(next) || next == '\u200d' || joined {
						continue
					}
					if cut != 0 && measure(word[:end]) > width {
						overflow = true
						break
					}
					cut = end
				}
				// Measure the whole remainder only if no earlier prefix exceeded
				// the width. Long replies must not be measured in full per line.
				if cut == 0 || (!overflow && measure(word) <= width) {
					break
				}
				lines = append(lines, word[:cut])
				word = word[cut:]
			}
			line = word
		}
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

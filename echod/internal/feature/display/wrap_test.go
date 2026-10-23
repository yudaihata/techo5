//go:build !dot

package display

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestWrapJapaneseReplyFitsWithoutLosingCharacters(t *testing.T) {
	text := strings.Repeat("日本語の返答が画面の端で見切れずに折り返されます。", 20)
	// Full-width text occupies two cells; Latin letters occupy one.
	measure := func(s string) int {
		w := 0
		for _, r := range s {
			if r > 127 {
				w += 2
			} else {
				w++
			}
		}
		return w
	}
	lines := wrapText(text, 35, measure)
	if len(lines) < 2 || strings.Join(lines, "") != text {
		t.Fatal("Japanese reply was not wrapped losslessly")
	}
	for _, line := range lines {
		if !utf8.ValidString(line) || measure(line) > 35 {
			t.Errorf("invalid or overflowing line: %q", line)
		}
	}
}

func TestWrapTextKeepsWordsAndParagraphs(t *testing.T) {
	for _, tc := range []struct {
		text  string
		width int
		want  []string
	}{
		{"one two three", 7, []string{"one two", "three"}},
		{"one\n\ntwo\r\nthree", 10, []string{"one", "", "two", "three"}},
		{"abcdefghij", 4, []string{"abcd", "efgh", "ij"}},
		{"日本語です", 2, []string{"日本", "語で", "す"}},
		{"日本語", 0, []string{"日", "本", "語"}},
		{"", 10, nil},
	} {
		if got := wrapText(tc.text, tc.width, utf8.RuneCountInString); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("wrapText(%q, %d) = %q, want %q", tc.text, tc.width, got, tc.want)
		}
	}
}

func TestWrapTextKeepsCombiningMarksWithTheirLetter(t *testing.T) {
	if got := wrapText("か\u3099き\u3099", 2, utf8.RuneCountInString); !reflect.DeepEqual(got, []string{"か\u3099", "き\u3099"}) {
		t.Errorf("split a combining mark: %q", got)
	}
}

func TestWrapLongReplyMeasuresOnlyLineSizedPrefixes(t *testing.T) {
	text := strings.Repeat("日本語の長い返答", 1000)
	longest := 0
	lines := wrapText(text, 30, func(s string) int {
		n := utf8.RuneCountInString(s)
		if n > longest {
			longest = n
		}
		return n
	})
	if strings.Join(lines, "") != text || longest > 31 {
		t.Errorf("lost text or repeatedly measured a long reply: longest measurement %d", longest)
	}
}

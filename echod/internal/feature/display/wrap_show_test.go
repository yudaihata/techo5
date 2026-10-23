//go:build !dot && !spot

package display

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestJapaneseReplyRendersWithinPanel(t *testing.T) {
	previewFont(t)
	reply := "日本語の返答を画面の幅に合わせて折り返します。単語の間に空白がなくても、右端で文章が見切れることはありません。長い返答は次の行に続きます。英数字のHome AssistantやESPHomeが文章の途中にあっても、文字を壊さずに表示します。"
	for _, size := range [][2]int{{960, 480}, {1280, 800}} {
		canvas := image.NewRGBA(image.Rect(0, 0, size[0], size[1]))
		r := newRenderer(canvas)
		for _, face := range []struct {
			name  string
			small bool
		}{{"body", false}, {"small", true}} {
			f := r.body
			if face.small {
				f = r.small
			}
			lines := r.wrap(f, reply, r.w-2*r.margin)
			joined := strings.Join(strings.Fields(strings.Join(lines, "")), "")
			if len(lines) < 2 || joined != strings.Join(strings.Fields(reply), "") {
				t.Fatalf("%v %s: reply was not wrapped losslessly", size, face.name)
			}
			for _, line := range lines {
				if !utf8.ValidString(line) || r.width(f, line) > r.w-2*r.margin {
					t.Errorf("%v %s: line overflows panel: %q", size, face.name, line)
				}
			}
		}
		r.draw(scene{now: time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC), phase: "lingering", heard: "日本語で説明してください", reply: reply, demo: true})
		if dir := os.Getenv("SHOW_PREVIEW"); dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, fmt.Sprintf("japanese-reply-%dx%d.png", size[0], size[1]))
			f, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(f, canvas)
			closeErr := f.Close()
			if err != nil {
				t.Fatal(err)
			}
			if closeErr != nil {
				t.Fatal(closeErr)
			}
		}
	}
}

//go:build !dot && !spot

package display

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/config"
	"github.com/HuskerMinion/techo5/echod/internal/feature/home"
	"github.com/HuskerMinion/techo5/echod/internal/lib/fonts"
)

func TestNotoPrimaryScreen(t *testing.T) {
	config.Use(filepath.Join(t.TempDir(), "state.json"))
	oldDir := fonts.Dir
	fonts.Dir = t.TempDir()
	fonts.Rescan()
	t.Cleanup(func() {
		fonts.Dir = oldDir
		fonts.Rescan()
		_ = config.Set().Screen().Font("")
		_ = config.Set().Screen().Language("")
		_ = config.Set().Screen().ClockStyle("")
	})
	if err := config.Set().Screen().Font(fonts.NotoSansCJKJP); err != nil {
		t.Fatal(err)
	}
	if err := config.Set().Screen().Language("ja"); err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 10, 23, 21, 45, 0, 0, time.UTC)
	for _, style := range []string{"", stylePhoto, styleOverlap} {
		_ = config.Set().Screen().ClockStyle(style)
		for _, reply := range []bool{false, true} {
			canvas := image.NewRGBA(image.Rect(0, 0, 1280, 800))
			r := newRenderer(canvas)
			s := scene{now: at, phase: "idle", weather: home.Weather{Condition: "rainy", Temp: "18°"}}
			name := style
			if name == "" {
				name = "classic"
			}
			if style == stylePhoto {
				s.slideshow = testPhoto("contrast", 1280, 800)
			}
			if reply {
				s.phase = "lingering"
				s.heard = "日本語と英数字を同じフォントで表示してください"
				s.reply = strings.Repeat("日本語の文章とHome Assistant、時刻21:45をNoto Sans CJK JPで表示しています。", 3)
				name += "-reply"
			}
			for _, line := range r.wrap(r.body, "日本語の文章と英数字Home Assistantを折り返します。", r.w-2*r.margin) {
				if r.width(r.body, line) > r.w-2*r.margin {
					t.Fatal("Noto text overflows")
				}
			}
			r.draw(s)
			if dir := os.Getenv("SHOW_PREVIEW"); dir != "" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				f, err := os.Create(filepath.Join(dir, "noto-"+name+"-show8.png"))
				if err != nil {
					t.Fatal(err)
				}
				err = png.Encode(f, canvas)
				f.Close()
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

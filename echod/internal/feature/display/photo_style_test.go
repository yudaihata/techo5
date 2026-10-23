//go:build !dot && !spot

package display

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/config"
	"github.com/HuskerMinion/techo5/echod/internal/feature/home"
)

func TestPhotoClock(t *testing.T) {
	previewFont(t)
	config.Use(filepath.Join(t.TempDir(), "state.json"))
	if err := config.Set().Screen().ClockStyle(stylePhoto); err != nil {
		t.Fatal(err)
	}
	if err := config.Set().Screen().Language("ja"); err != nil {
		t.Fatal(err)
	}
	if err := config.Set().Screen().Font("inter"); err != nil {
		t.Fatal(err)
	}
	old24 := clock24.Load()
	clock24.Store(true)
	t.Cleanup(func() {
		clock24.Store(old24)
		_ = config.Set().Screen().ClockStyle("")
		_ = config.Set().Screen().Language("")
		_ = config.Set().Screen().Font("")
	})
	if clockStyleOptions()[clockStyleIndex()] != "Photo" {
		t.Fatal("Photo is not selectable")
	}
	for _, size := range [][2]int{{960, 480}, {1280, 800}} {
		for _, kind := range []string{"snow", "night", "contrast"} {
			for _, mode := range []string{"background", "screensaver"} {
				photo := testPhoto(kind, size[0], size[1])
				img := image.NewRGBA(photo.Rect)
				r := newRenderer(img)
				s := scene{now: time.Date(2026, 10, 23, 13, 39, 0, 0, time.UTC), phase: "idle", weather: home.Weather{Condition: "cloudy", Temp: "31°"}}
				if mode == "background" {
					s.slideshow = photo
				} else {
					s.slideshowScreensaver = photo
				}
				r.draw(s)
				// Photo keeps the photo's center intact, and never recolors its date/weather gold.
				p := image.Pt(size[0]/2, size[1]/3)
				if img.RGBAAt(p.X, p.Y) != photo.RGBAAt(p.X, p.Y) {
					t.Errorf("%v %s %s: photograph was globally washed", size, kind, mode)
				}
				white := 0
				for y := size[1] / 2; y < size[1]; y++ {
					for x := 0; x < size[0]; x++ {
						c := img.RGBAAt(x, y)
						if c == photoGold {
							t.Fatalf("%v %s %s: gold text survived", size, kind, mode)
						}
						if c.R == 255 && c.G == 255 && c.B == 255 {
							white++
						}
					}
				}
				if white < 50 {
					t.Errorf("%v %s %s: white clock text missing", size, kind, mode)
				}
				if dir := os.Getenv("SHOW_PREVIEW"); dir != "" {
					if err := os.MkdirAll(dir, 0755); err != nil {
						t.Fatal(err)
					}
					path := filepath.Join(dir, fmt.Sprintf("photo-%s-%s-%dx%d.png", mode, kind, size[0], size[1]))
					f, err := os.Create(path)
					if err != nil {
						t.Fatal(err)
					}
					err = png.Encode(f, img)
					f.Close()
					if err != nil {
						t.Fatal(err)
					}
				}
			}
		}
	}
}

func TestPhotoClockDeviceSelection(t *testing.T) {
	config.Use(filepath.Join(t.TempDir(), "state.json"))
	d := Get()
	if !slices.Contains(d.clockStyleSel.Options, "Photo") {
		t.Fatal("the device's already-created API select does not offer Photo")
	}
	d.clockStyleSel.OnCommand("Photo")
	if config.Get().Screen.ClockStyle != stylePhoto {
		t.Fatal("the actual device select did not persist Photo")
	}
	d.clockStyleSel.OnCommand("Classic")
}

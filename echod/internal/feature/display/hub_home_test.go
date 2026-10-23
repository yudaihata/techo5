//go:build !dot && !spot

package display

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/config"
	"github.com/HuskerMinion/techo5/echod/internal/feature/home"
	"github.com/HuskerMinion/techo5/echod/internal/feature/media"
	"github.com/HuskerMinion/techo5/echod/internal/feature/phone"
	"github.com/HuskerMinion/techo5/echod/internal/feature/voice"
	"github.com/HuskerMinion/techo5/echod/internal/hardware/touch"
	"github.com/HuskerMinion/techo5/echod/internal/lib/fonts"
	"github.com/HuskerMinion/techo5/echod/internal/lib/mdi"
)

func hubTestDisplay(t *testing.T) *Display {
	t.Helper()
	config.Use(filepath.Join(t.TempDir(), "state.json"))
	t.Cleanup(func() {
		_ = config.Set().Screen().NestHome(false)
		touch.Get().SetFollowArea(image.Rectangle{})
	})
	d := &Display{on: true, ceiling: 80, clockUp: true, view: voice.State{Phase: "idle"}, r: newRenderer(image.NewRGBA(image.Rect(0, 0, 1280, 800))), poke: make(chan struct{}, 1)}
	d.hubSwitch = hubHomeSwitch(d)
	if err := config.Set().Screen().NestHome(true); err != nil {
		t.Fatal(err)
	}
	d.hub.shown = true
	return d
}

func TestHubQuickSettingsGesture(t *testing.T) {
	d := hubTestDisplay(t)
	volume := media.Get().Volume()
	if !d.hubGesture(touch.Gesture{Kind: touch.SwipeUp, X: 600, Y: 798}) || d.hub.panel != hubMenu {
		t.Fatal("bottom swipe did not open menu")
	}
	for range 15 {
		d.hubGesture(touch.Gesture{Kind: touch.SwipeUp, X: 600, Y: 798})
	}
	if media.Get().Volume() != volume {
		t.Fatal("opening finger also changed volume")
	}
	d.hub.opened = time.Now().Add(-time.Second)
	b := d.r.hubPanelBox(false)
	d.hubGesture(touch.Gesture{Kind: touch.Tap, X: b.Min.X + b.Dx()/10, Y: b.Min.Y + b.Dy()/2})
	if d.hub.panel != hubBrightness {
		t.Fatal("brightness button did not change panel")
	}
	d.hubGesture(touch.Gesture{Kind: touch.Tap, X: 600, Y: 200})
	if d.hub.closing.IsZero() {
		t.Fatal("background tap did not dismiss")
	}
	s := scene{now: time.Now().Add(time.Second), phase: "idle"}
	d.hubScene(&s)
	if s.hub.panel != hubClosed {
		t.Fatal("dismissed panel remains")
	}
	_ = config.Set().Screen().NestHome(false)
	if d.hubGesture(touch.Gesture{Kind: touch.SwipeUp, Y: 798}) {
		t.Fatal("disabled UI intercepted a standard gesture")
	}
}

func TestHubAlarmVolumeExpansion(t *testing.T) {
	d := hubTestDisplay(t)
	d.hub.panel = hubVolume
	d.hub.opened = time.Now().Add(-time.Second)
	d.hub.touched = time.Now()
	b := d.r.hubPanelBox(false)
	d.hubGesture(touch.Gesture{Kind: touch.Tap, X: b.Max.X - 25, Y: b.Min.Y + b.Dy()/2})
	if !d.hub.rings {
		t.Fatal("volume number did not expand alarm volume")
	}
	d.hub.ringAt = time.Now().Add(-time.Second)
	d.hubGesture(touch.Gesture{Kind: touch.Tap, X: b.Max.X - 25, Y: b.Min.Y + b.Dy()/2})
	if d.hub.ringClosing.IsZero() {
		t.Fatal("volume number did not start collapsing alarm volume")
	}
	scene := scene{now: time.Now().Add(time.Second), phase: "idle"}
	d.hubScene(&scene)
	if d.hub.rings {
		t.Fatal("alarm volume did not finish collapsing")
	}
}

func TestHubTopAndRightOpenSameSmartHome(t *testing.T) {
	for _, g := range []touch.Gesture{{Kind: touch.SwipeDown, X: 600, Y: 1}, {Kind: touch.SwipeLeft, X: 1278, Y: 400}} {
		d := hubTestDisplay(t)
		if !d.hubGesture(g) || !d.hub.smart {
			t.Fatal("edge did not open smart home", g)
		}
		want := image.Pt(-1, 0)
		if g.Kind == touch.SwipeDown {
			want = image.Pt(0, 1)
		}
		if d.hub.pageDir != want {
			t.Fatal("page moved against swipe", d.hub.pageDir)
		}
		d.hub.pageAt = time.Now().Add(-time.Second)
		d.hubGesture(touch.Gesture{Kind: touch.SwipeRight, X: 10, Y: 400})
		if d.hub.smart {
			t.Fatal("return swipe did not close smart home")
		}
	}
}

func TestHubSliderStepsAndHitBounds(t *testing.T) {
	for _, size := range []image.Point{image.Pt(960, 480), image.Pt(1280, 800)} {
		r := newRenderer(image.NewRGBA(image.Rectangle{Max: size}))
		if r.hubPanelBox(false).Overlaps(r.hubPanelBox(true)) {
			t.Fatal("alarm and media panels overlap")
		}
		for _, upper := range []bool{false, true} {
			b := r.hubDotsBox(upper)
			for i := range 10 {
				x := b.Min.X + int((float64(i)+0.5)*float64(b.Dx())/10)
				if got := r.hubDotStep(x, upper); got != i+1 {
					t.Fatalf("dot %d maps to %d", i+1, got)
				}
			}
			if r.hubDotStep(b.Min.X-50, upper) != 0 || r.hubDotStep(b.Max.X+50, upper) != 10 {
				t.Fatal("slider did not clamp")
			}
		}
	}
	for n := range 11 {
		if hubToTen(hubFromTen(n, media.VolumeSteps), media.VolumeSteps) != n {
			t.Fatal("volume step lost", n)
		}
	}
	for _, name := range []string{"white-balance-sunny", "volume-high", "minus-circle-outline", "alarm-plus", "cog-outline", "microphone-off", "home-outline", "alarm", "chevron-up", "chevron-down"} {
		if _, ok := mdi.Rune(name); !ok {
			t.Fatal("missing control icon", name)
		}
	}
}

func TestHubUrgentPagesPreemptControls(t *testing.T) {
	cases := []struct {
		name string
		set  func(*scene)
	}{
		{"ring", func(s *scene) { s.ring.timer = "timer" }},
		{"call", func(s *scene) { s.call.Phase = phone.Ringing }},
		{"voice", func(s *scene) { s.phase = "listening" }},
		{"voice-cue", func(s *scene) { s.cue = "listening" }},
		{"pin", func(s *scene) { s.pin.open = true }},
		{"settings", func(s *scene) { s.showSheet = true }},
		{"notification", func(s *scene) { s.showNotification = true }},
		{"update", func(s *scene) { s.updateActive = true }},
		{"video", func(s *scene) { s.showVideo = true }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := hubTestDisplay(t)
			d.hub.panel = hubMenu
			d.hub.opened = time.Now()
			d.hub.touched = time.Now()
			s := scene{now: time.Now(), phase: "idle"}
			c.set(&s)
			d.hubScene(&s)
			if s.hub.shown || d.hub.panel != hubClosed {
				t.Fatal("controls cover urgent page")
			}
		})
	}
}

func TestHubSlideDirection(t *testing.T) {
	b := image.Rect(0, 0, 100, 60)
	from, to, dst := image.NewRGBA(b), image.NewRGBA(b), image.NewRGBA(b)
	red, blue := color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255}
	draw.Draw(from, b, image.NewUniform(red), image.Point{}, draw.Src)
	draw.Draw(to, b, image.NewUniform(blue), image.Point{}, draw.Src)
	for _, dir := range []image.Point{image.Pt(1, 0), image.Pt(-1, 0), image.Pt(0, 1), image.Pt(0, -1)} {
		hubSlide(dst, from, to, dir, 0)
		if dst.RGBAAt(50, 30) != red {
			t.Fatal("slide starts on wrong page")
		}
		hubSlide(dst, from, to, dir, 0.5)
		newAt, oldAt := image.Pt(20, 20), image.Pt(80, 40)
		if dir.X < 0 || dir.Y < 0 {
			newAt, oldAt = oldAt, newAt
		}
		if dst.RGBAAt(newAt.X, newAt.Y) != blue || dst.RGBAAt(oldAt.X, oldAt.Y) != red {
			t.Fatal("slide moves against direction", dir)
		}
		hubSlide(dst, from, to, dir, 1)
		if dst.RGBAAt(50, 30) != blue {
			t.Fatal("slide does not finish")
		}
	}
}

func TestHubHomePreviews(t *testing.T) {
	hubTestDisplay(t)
	_ = config.Set().Screen().ClockStyle(stylePhoto)
	_ = config.Set().Screen().Font(fonts.NotoSansCJKJP)
	_ = config.Set().Screen().Language("ja")
	t.Cleanup(func() {
		_ = config.Set().Screen().ClockStyle("")
		_ = config.Set().Screen().Font("")
		_ = config.Set().Screen().Language("")
	})
	at := time.Date(2026, 10, 23, 21, 4, 0, 0, time.UTC)
	for _, sample := range []struct {
		name string
		v    hubView
	}{
		{"menu", hubView{panel: hubMenu}},
		{"brightness", hubView{panel: hubBrightness, brightness: 8}},
		{"brightness-auto", hubView{panel: hubBrightness, brightness: 8, automatic: true}},
		{"volume", hubView{panel: hubVolume, volume: 5}},
		{"volume-alarm", hubView{panel: hubVolume, volume: 5, rings: true, alarm: 3, ringReveal: 1}},
		{"mic-muted", hubView{panel: hubMenu, muted: true}},
		{"night", hubView{night: true}},
		{"smart-unconfigured", hubView{empty: true}},
	} {
		for _, size := range []image.Point{image.Pt(1280, 800), image.Pt(960, 480)} {
			canvas := image.NewRGBA(image.Rectangle{Max: size})
			r := newRenderer(canvas)
			v := sample.v
			v.enabled = true
			v.shown = true
			v.reveal = 1
			s := scene{now: at, phase: "idle", slideshow: testPhoto("contrast", size.X, size.Y), weather: home.Weather{Condition: "rainy", Temp: "18°"}, hub: v}
			r.draw(s)
			if dir := os.Getenv("SHOW_PREVIEW"); dir != "" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				model := "show8"
				if size.X == 960 {
					model = "show5"
				}
				f, err := os.Create(filepath.Join(dir, "hub-"+sample.name+"-"+model+".png"))
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

func TestHubHardwareMuteButton(t *testing.T) {
	d := hubTestDisplay(t)
	d.hub.panel = hubMenu
	d.hub.opened = time.Now().Add(-time.Second)
	calls := 0
	previous := flipHubMicrophone
	flipHubMicrophone = func() { calls++ }
	t.Cleanup(func() { flipHubMicrophone = previous })
	b := d.r.hubPanelBox(false)
	d.hubGesture(touch.Gesture{Kind: touch.Tap, X: b.Min.X + b.Dx()/2, Y: b.Min.Y + b.Dy()/2})
	if calls != 1 {
		t.Fatal("button did not use hardware microphone switch")
	}
}

func TestHubGlassPreservesThePhoto(t *testing.T) {
	r := newRenderer(image.NewRGBA(image.Rect(0, 0, 1280, 800)))
	a, b := image.Pt(300, 700), image.Pt(800, 700)
	r.dst.SetRGBA(a.X, a.Y, color.RGBA{255, 255, 255, 255})
	r.dst.SetRGBA(b.X, b.Y, color.RGBA{0, 0, 0, 255})
	r.hubGlass(r.hubPanelBox(false))
	if r.dst.RGBAAt(a.X, a.Y) == r.dst.RGBAAt(b.X, b.Y) {
		t.Fatal("quick panel replaced photo with opaque fill")
	}
}

func TestHubModeSwitchRestoresStandardControls(t *testing.T) {
	d := hubTestDisplay(t)
	d.hubSwitch.OnCommand(true)
	if !d.hubSwitch.Get() || !config.Get().Screen.NestHome {
		t.Fatal("Nest home choice did not persist")
	}
	d.hub.panel = hubVolume
	d.hubSwitch.OnCommand(false)
	if d.hubSwitch.Get() || config.Get().Screen.NestHome || d.hub.panel != hubClosed {
		t.Fatal("standard controls were not restored")
	}
}

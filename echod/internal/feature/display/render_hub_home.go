//go:build !dot && !spot

package display

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
)

// Only the frame goroutine owns these snapshots. A dashboard is rendered once per slide.
type hubAnimation struct {
	token    uint64
	from, to *image.RGBA
	mask     *image.Alpha
}

func hubCopy(src *image.RGBA) *image.RGBA {
	out := image.NewRGBA(src.Bounds())
	draw.Draw(out, out.Bounds(), src, src.Bounds().Min, draw.Src)
	return out
}
func (r *renderer) drawHubFrame(s scene) {
	v := s.hub
	a := &r.hubAnimation
	if !v.enabled || !v.shown || v.pageToken == 0 || s.now.Sub(v.pageAt) >= hubDuration {
		a.from, a.to = nil, nil
		a.token = v.pageToken
		r.drawScene(s)
		return
	}
	if a.token != v.pageToken || a.to == nil {
		a.from = hubCopy(r.dst)
		r.drawScene(s)
		a.to = hubCopy(r.dst)
		a.token = v.pageToken
	}
	hubSlide(r.dst, a.from, a.to, v.pageDir, hubEase(hubProgress(s.now, v.pageAt)))
}

// Both pages move with the swipe; the incoming page starts beyond the opposite edge.
func hubSlide(dst, from, to *image.RGBA, dir image.Point, progress float64) {
	b := dst.Bounds()
	w, h := b.Dx(), b.Dy()
	delta := image.Pt(int(math.Round(float64(dir.X*w)*progress)), int(math.Round(float64(dir.Y*h)*progress)))
	incoming := delta.Sub(image.Pt(dir.X*w, dir.Y*h))
	draw.Draw(dst, b, image.Black, image.Point{}, draw.Src)
	draw.Draw(dst, b.Add(delta), from, b.Min, draw.Src)
	draw.Draw(dst, b.Add(incoming), to, b.Min, draw.Src)
}
func (r *renderer) hubPanelBox(upper bool) image.Rectangle {
	margin := max(16, r.w/32)
	height := r.h / 5
	bottom := r.h - max(16, r.h/30)
	if upper {
		bottom -= height + max(12, r.h/35)
	}
	return image.Rect(margin, bottom-height, r.w-margin, bottom)
}
func (r *renderer) hubDotsBox(upper bool) image.Rectangle {
	b := r.hubPanelBox(upper)
	return image.Rect(b.Min.X+b.Dx()*13/100, b.Min.Y, b.Max.X-b.Dx()*13/100, b.Max.Y)
}
func (r *renderer) hubDotStep(x int, upper bool) int {
	b := r.hubDotsBox(upper)
	gap := float64(b.Dx()) / 10
	return min(max(int(math.Round((float64(x-b.Min.X)-gap/2)/gap))+1, 0), 10)
}

var hubWhite = color.RGBA{245, 247, 250, 255}
var hubPanelInk = color.RGBA{65, 70, 80, 220} // premultiplied translucent slate

// The settings painter's roundFill is opaque. Here a cached alpha mask composites the panel over
// the existing photo and clock, including its rounded corners, without darkening the whole screen.
func (r *renderer) hubGlass(b image.Rectangle) {
	mask := r.hubAnimation.mask
	if mask == nil || mask.Bounds().Size() != b.Size() {
		mask = image.NewAlpha(image.Rectangle{Max: b.Size()})
		rad := float64(r.h / 45)
		for y := 0; y < b.Dy(); y++ {
			for x := 0; x < b.Dx(); x++ {
				a := clamp01(0.5 - rrDist(float64(x)+0.5, float64(y)+0.5, 0, 0, float64(b.Dx()), float64(b.Dy()), rad))
				mask.SetAlpha(x, y, color.Alpha{A: uint8(255*a + 0.5)})
			}
		}
		r.hubAnimation.mask = mask
	}
	draw.DrawMask(r.dst, b, image.NewUniform(hubPanelInk), image.Point{}, mask, image.Point{}, draw.Over)
}

func (r *renderer) hubOverlay(s scene) {
	v := s.hub
	if !v.enabled || !v.shown {
		return
	}
	if v.night {
		draw.Draw(r.dst, r.dst.Rect, image.Black, image.Point{}, draw.Src)
		text := clockHM(s.now)
		face := r.textFace(false, 148)
		r.text(face, text, (r.w-r.width(face, text))/2, r.h/2+r.s(48), hubWhite)
		return
	}
	if v.empty {
		draw.Draw(r.dst, r.dst.Rect, image.NewUniform(color.RGBA{243, 245, 248, 255}), image.Point{}, draw.Src)
		ink := color.RGBA{51, 57, 66, 255}
		title, sub, back, settings := "Smart home", "Choose a dashboard in settings", "‹ Home", "Settings"
		if screenLang() == "ja" {
			title, sub, back, settings = "スマートホーム", "設定からダッシュボードを選択してください", "‹ ホーム", "本体設定"
		}
		face := r.textFace(false, 30)
		r.text(face, back, r.s(28), r.s(54), ink)
		r.mdiIconIn("home-outline", image.Rect(0, r.h/5, r.w, r.h/2), 64, ink)
		r.text(face, title, (r.w-r.width(face, title))/2, r.h/2, ink)
		small := r.textFace(false, 18)
		r.text(small, sub, (r.w-r.width(small, sub))/2, r.h/2+r.s(48), ink)
		r.text(small, settings, (r.w-r.width(small, settings))/2, r.h*3/4, ink)
	}
	if v.muted {
		r.mdiIconIn("microphone-off", image.Rect(r.w-r.s(48), r.s(10), r.w-r.s(10), r.s(48)), 24, color.RGBA{255, 166, 105, 255})
	}
	if v.panel == hubClosed {
		return
	}
	if v.muted && v.micHint {
		hint := "Unmute with the microphone button"
		if screenLang() == "ja" {
			hint = "マイクの解除は本体のボタンで操作"
		}
		f := r.textFace(false, 18)
		b := r.hubPanelBox(false)
		textBox := image.Rect((r.w-r.width(f, hint))/2-r.s(12), b.Min.Y-r.s(48), (r.w+r.width(f, hint))/2+r.s(12), b.Min.Y-r.s(8))
		r.roundFill(textBox, r.sf(10), color.RGBA{40, 44, 51, 255}, color.RGBA{40, 44, 51, 255})
		r.text(f, hint, textBox.Min.X+r.s(12), textBox.Max.Y-r.s(12), hubWhite)
	}

	offset := int(float64(r.h-r.hubPanelBox(false).Min.Y) * (1 - v.reveal))
	box := r.hubPanelBox(false).Add(image.Pt(0, offset))
	// Draw the alarm row behind the media row while it slides upward out of it.
	if v.panel == hubVolume && v.rings {
		upper := r.hubPanelBox(true)
		upper = upper.Add(image.Pt(0, offset+int(float64(upper.Dy()+r.h/35)*(1-v.ringReveal))))
		r.hubGlass(upper)
		r.hubSlider(upper, v.alarm, "alarm", false, false, false)
	}
	r.hubGlass(box)
	if v.panel == hubMenu {
		names := []string{"white-balance-sunny", "volume-high", "minus-circle-outline", "alarm-plus", "cog-outline"}
		for i, name := range names {
			b := image.Rect(box.Min.X+box.Dx()*i/5, box.Min.Y, box.Min.X+box.Dx()*(i+1)/5, box.Max.Y)
			ink := hubWhite
			if i == 2 && v.muted {
				ink = color.RGBA{255, 166, 105, 255}
			}
			r.mdiIconIn(name, b, 42, ink)
		}
		return
	}
	n, icon := v.brightness, "white-balance-sunny"
	if v.panel == hubVolume {
		n, icon = v.volume, "volume-high"
	}
	r.hubSlider(box, n, icon, v.automatic && v.panel == hubBrightness, v.panel == hubVolume, v.rings)
}
func (r *renderer) hubSlider(box image.Rectangle, n int, icon string, automatic, expand, expanded bool) {
	iconBox := image.Rect(box.Min.X, box.Min.Y, box.Min.X+box.Dx()*13/100, box.Max.Y)
	r.mdiIconIn(icon, iconBox, 42, hubWhite)
	if automatic {
		center := image.Pt((iconBox.Min.X+iconBox.Max.X)/2, (box.Min.Y+box.Max.Y)/2)
		r.disc(center.X, center.Y, r.s(10), color.RGBA{75, 81, 93, 255})
		f := r.textFace(true, 15)
		r.text(f, "A", center.X-r.width(f, "A")/2, center.Y+r.s(6), hubWhite)
	}
	left, right := box.Min.X+box.Dx()*13/100, box.Max.X-box.Dx()*13/100
	gap := float64(right-left) / 10
	cy := (box.Min.Y + box.Max.Y) / 2
	for i := range 10 {
		ink := color.RGBA{142, 149, 162, 255}
		if i < n {
			ink = hubWhite
		}
		r.disc(left+int((float64(i)+0.5)*gap), cy, max(4, r.w/160), ink)
	}
	f := r.textFace(false, 36)
	number := fmt.Sprint(n)
	x := right + (box.Max.X-right-r.width(f, number))/2
	if expand {
		x -= r.s(6)
	}
	r.text(f, number, x, cy+r.s(13), hubWhite)
	if expand {
		name := "chevron-up"
		if expanded {
			name = "chevron-down"
		}
		r.mdiIconIn(name, image.Rect(box.Max.X-r.s(33), box.Min.Y, box.Max.X-r.s(8), box.Max.Y), 22, hubWhite)
	}
}

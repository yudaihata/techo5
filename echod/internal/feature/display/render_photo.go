//go:build !dot && !spot

package display

import (
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"

	"github.com/HuskerMinion/techo5/echod/internal/lib/locale"
)

const stylePhoto = "photo"

// Show-only registration keeps the round Spot's choices and renderers unchanged.
func init() {
	clockStyles = append(clockStyles, struct{ label, value string }{"Photo", stylePhoto})
}

func photoWash() uint8 {
	if clockStyle() == stylePhoto {
		return 0
	}
	return slideshowWash
}

// photoStyle groups a quiet date, regular-weight time and monochrome weather at the lower left.
// It uses the shared style layout's bottom so music, timers and next-up cards keep their space.
func (r *renderer) photoStyle(s scene, box image.Rectangle) {
	x, base := r.margin, box.Max.Y-r.s(6)
	if s.callButton {
		x = max(x, r.callButtonRect().Max.X+r.s(20))
	}
	photo := s.slideshow != nil || s.slideshowScreensaver != nil
	ink := cream
	if photo {
		ink = color.RGBA{255, 255, 255, 255}
	}
	timeFace, dateFace, weatherFace := r.styleFace(false, 84), r.styleFace(false, 23), r.styleFace(false, 28)
	timeText := clockHM(s.now)
	gap, iconSize := r.s(20), r.s(27)
	suffix := clockSuffix(s.now)
	icon := s.weather.Condition != "" && conditionWords(s.weather.Condition) != ""
	weatherWidth := r.width(weatherFace, s.weather.Temp)
	if icon {
		weatherWidth += iconSize + r.s(6)
	}
	suffixWidth := 0
	if suffix != "" {
		suffixWidth = r.width(dateFace, suffix) + r.s(12)
	}
	room := r.w - r.margin - x - weatherWidth - gap - suffixWidth
	if w := r.width(timeFace, timeText); w > room {
		timeFace = r.styleFace(false, max(40, 84*room/max(w, 1)))
	}
	timeWidth := r.width(timeFace, timeText)
	dateBase := base - capHeight(timeFace) - r.s(15)
	if photo {
		// A soft, continuous shade at the foot; no global wash or pills around individual lines.
		top := max(box.Min.Y, dateBase-r.s(40))
		for y := top; y < r.h; y++ {
			a := uint8(min(76, max(0, (y-top)*76/max(r.h-top, 1))))
			draw.Draw(r.dst, image.Rect(0, y, r.w, y+1), image.NewUniform(color.RGBA{0, 0, 0, a}), image.Point{}, draw.Over)
		}
	}
	text := func(f font.Face, line string, tx, baseline int) {
		if photo {
			r.text(f, line, tx+r.s(1), baseline+r.s(2), chosenColor(color.RGBA{0, 0, 0, 130}))
		}
		r.text(f, line, tx, baseline, chosenColor(ink))
	}
	date := locale.LongDate(s.now, screenLang())
	if screenLang() == "ja" {
		date = s.now.Format("1月2日") + "(" + locale.ShortWeekday(s.now, "ja") + ")"
	}
	date = r.clipTo(dateFace, date, r.w-r.margin-x)
	text(dateFace, date, x, dateBase)
	r.setDateAt(image.Rect(x, dateBase-r.s(28), x+r.width(dateFace, date), dateBase+r.s(8)))
	text(timeFace, timeText, x, base)
	wx := x + timeWidth + r.s(12)
	if suffix != "" {
		text(dateFace, suffix, wx, base-r.s(4))
		wx += suffixWidth
	}
	wx += gap - r.s(12)
	weatherTop := base - capHeight(weatherFace) - r.s(3)
	if icon {
		name, _ := weatherGlyph(s.weather.Condition)
		iconBox := image.Rect(wx, weatherTop, wx+iconSize, weatherTop+iconSize)
		if photo {
			r.mdiIconIn(name, iconBox.Add(image.Pt(r.s(1), r.s(2))), 27, color.RGBA{0, 0, 0, 130})
		}
		r.mdiIconIn(name, iconBox, 27, ink)
		wx += iconSize + r.s(6)
	}
	if s.weather.Temp != "" {
		text(weatherFace, s.weather.Temp, wx, base-r.s(3))
	}
	r.setWeatherAt(image.Rect(x+timeWidth+suffixWidth, weatherTop-r.s(10), wx+r.width(weatherFace, s.weather.Temp), base+r.s(10)).Inset(-r.s(8)))
}

//go:build !dot && !spot

package display

import (
	"image"
	"log/slog"
	"math"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/config"
	"github.com/HuskerMinion/techo5/echod/internal/feature/alarm"
	"github.com/HuskerMinion/techo5/echod/internal/feature/dashboard"
	"github.com/HuskerMinion/techo5/echod/internal/feature/media"
	"github.com/HuskerMinion/techo5/echod/internal/feature/mute"
	"github.com/HuskerMinion/techo5/echod/internal/feature/phone"
	"github.com/HuskerMinion/techo5/echod/internal/feature/ring"
	"github.com/HuskerMinion/techo5/echod/internal/hardware/touch"
	esphome "github.com/ygelfand/go-esphome-device"
)

const hubDuration = 280 * time.Millisecond
const hubIdle = 30 * time.Second

type hubPanel int

const (
	hubClosed hubPanel = iota
	hubMenu
	hubBrightness
	hubVolume
)

type hubState struct {
	panel                    hubPanel
	opened, closing, touched time.Time
	rings                    bool
	ringAt                   time.Time
	ringClosing              time.Time
	micHintUntil             time.Time
	night, smart, shown      bool
	slider                   int
	pageToken                uint64
	pageDir                  image.Point
	pageAt                   time.Time
	openingContact           uint64
}

type hubView struct {
	enabled, shown, night, empty, muted, automatic, rings, moving bool
	panel                                                         hubPanel
	micHint                                                       bool
	reveal, ringReveal                                            float64
	brightness, volume, alarm                                     int
	pageToken                                                     uint64
	pageDir                                                       image.Point
	pageAt                                                        time.Time
}

func hubHomeSwitch(d *Display) *esphome.Switch {
	s := &esphome.Switch{Base: esphome.Base{ObjectID: "screen_nest_home", Name: "Nest-style home", Icon: "mdi:gesture-swipe", Category: esphome.CategoryConfig}}
	s.OnCommand = d.setHubHome
	return s
}

func (d *Display) setHubHome(on bool) {
	if err := config.Set().Screen().NestHome(on); err != nil {
		slog.Error("saving Nest-style home failed", "err", err)
		return
	}
	d.mu.Lock()
	wasNight := d.hub.night
	d.hub = hubState{}
	d.mu.Unlock()
	if d.hubSwitch != nil {
		d.hubSwitch.Set(on)
	}
	touch.Get().SetFollowArea(image.Rectangle{})
	if wasNight {
		d.relight(true)
	}
	d.wake()
}

func hubEase(t float64) float64 { t = min(max(t, 0), 1); return 1 - math.Pow(1-t, 3) }
func hubProgress(now, at time.Time) float64 {
	return min(max(float64(now.Sub(at))/float64(hubDuration), 0), 1)
}
func (h *hubState) page(dir image.Point, now time.Time) {
	h.pageToken++
	h.pageDir = dir
	h.pageAt = now
}

// Only a clock, or the smart-home page itself, may carry these controls. Alerts, settings,
// conversations, calls and setup retain their existing priorities and input handling.
func hubEligible(s *scene) bool {
	return s.phase == "idle" && s.cue == "" && !s.ring.any() && s.call.Phase == phone.Idle && !s.setupAsking && !s.showVideoAsk && !s.showVideo &&
		!s.pin.open && !s.showWifi && !s.bt.Pairing && !s.showSheet && !s.showDrawer && !s.showDeck && !s.showWeather &&
		!s.showCalendar && !s.showCamera && !s.showAlert && !s.showNotification && !s.showReminder && s.popup == nil &&
		!s.showAnnouncement && !s.announceRecording && !s.updateActive && !s.nowPlaying && s.sunrise == 0
}

func (d *Display) hubScene(s *scene) {
	enabled := config.Get().Screen.NestHome
	now := s.now
	d.mu.Lock()
	h := &d.hub
	if !h.closing.IsZero() && now.Sub(h.closing) >= hubDuration {
		h.panel = hubClosed
		h.closing = time.Time{}
		h.slider = 0
	}
	if !h.ringClosing.IsZero() && now.Sub(h.ringClosing) >= hubDuration {
		h.rings = false
		h.ringClosing = time.Time{}
	}
	if h.panel != hubClosed && h.closing.IsZero() && now.Sub(h.touched) > hubIdle {
		h.closing = now
	}
	eligible := enabled && hubEligible(s)
	wasNight := h.night
	if !eligible {
		h.panel = hubClosed
		h.night = false
		h.slider = 0
		h.rings = false
	}

	h.shown = eligible && (!s.showDash || h.smart)
	v := hubView{enabled: enabled, shown: h.shown, panel: h.panel, night: h.night, empty: h.smart && !s.showDash,
		micHint: now.Before(h.micHintUntil), muted: s.muted, automatic: d.autoOn, brightness: max(1, (d.ceiling+5)/10),
		volume: hubToTen(media.Get().Volume(), media.VolumeSteps), alarm: hubToTen(ring.Level(), media.VolumeSteps),
		rings: h.rings, pageToken: h.pageToken, pageDir: h.pageDir, pageAt: h.pageAt}
	if h.panel != hubClosed {
		v.reveal = hubEase(hubProgress(now, h.opened))
		if !h.closing.IsZero() {
			v.reveal = 1 - hubEase(hubProgress(now, h.closing))
		}
		v.moving = now.Sub(h.opened) < hubDuration || !h.closing.IsZero()
	}
	if h.rings {
		v.ringReveal = hubEase(hubProgress(now, h.ringAt))
		if !h.ringClosing.IsZero() {
			v.ringReveal = 1 - hubEase(hubProgress(now, h.ringClosing))
			v.moving = true
		}
		v.moving = v.moving || now.Sub(h.ringAt) < hubDuration
	}
	if h.shown && h.pageToken != 0 && now.Sub(h.pageAt) < hubDuration {
		v.moving = true
	}
	d.mu.Unlock()
	s.hub = v
	if !eligible {
		s.hub.pageToken = 0
	}
	if wasNight && !v.night {
		d.relight(true)
	}
	area := image.Rectangle{}
	if v.shown && v.panel >= hubBrightness && d.r != nil {
		area = d.r.hubPanelBox(false)
		if v.rings {
			area = area.Union(d.r.hubPanelBox(true))
		}
	}
	touch.Get().SetFollowArea(area)
	if v.shown && v.panel != hubClosed {
		s.showVolume = false
	}
}

func hubToTen(n, maxN int) int   { return min(max((n*10+maxN/2)/maxN, 0), 10) }
func hubFromTen(n, maxN int) int { return (min(max(n, 0), 10)*maxN + 5) / 10 }

func (d *Display) hubClosePanel(now time.Time) {
	d.mu.Lock()
	d.hub.closing = now
	d.hub.slider = 0
	d.mu.Unlock()
	d.wake()
}

// hubGesture is intentionally below urgent pages and above the old clock gestures. Opening a
// panel consumes every remaining notch of that same finger; it never also changes the volume.
func (d *Display) hubGesture(g touch.Gesture) bool {
	if !config.Get().Screen.NestHome || d.r == nil {
		return false
	}
	now := time.Now()
	d.mu.Lock()
	h := d.hub
	d.mu.Unlock()
	if !h.shown {
		return false
	}
	contact, down := touch.Get().Contact()
	if down && contact != 0 && contact == h.openingContact {
		return true
	}
	if h.pageToken != 0 && now.Sub(h.pageAt) < hubDuration {
		return true
	}
	if h.night {
		if g.Kind == touch.Tap || g.Kind == touch.SwipeLeft {
			d.mu.Lock()
			d.hub.night = false
			d.hub.page(image.Pt(-1, 0), now)
			d.mu.Unlock()
			d.relight(true)
			d.wake()
		}
		return true
	}
	if h.panel != hubClosed {
		if !h.closing.IsZero() {
			return true
		}
		if now.Sub(h.opened) < hubDuration || (h.rings && now.Sub(h.ringAt) < hubDuration) || !h.ringClosing.IsZero() {
			return true
		}
		d.mu.Lock()
		d.hub.touched = now
		d.mu.Unlock()
		switch g.Kind {
		case touch.SwipeDown:
			d.hubClosePanel(now)
		case touch.Tap:
			p := image.Pt(g.X, g.Y)
			box := d.r.hubPanelBox(false)
			if h.panel == hubMenu {
				if !p.In(box) {
					d.hubClosePanel(now)
					return true
				}
				switch (g.X - box.Min.X) * 5 / box.Dx() {
				case 0:
					d.mu.Lock()
					d.hub.panel = hubBrightness
					d.mu.Unlock()
				case 1:
					d.mu.Lock()
					d.hub.panel = hubVolume
					d.mu.Unlock()
				case 2:
					flipHubMicrophone()
					d.mu.Lock()
					d.hub.micHintUntil = now.Add(4 * time.Second)
					d.mu.Unlock()
				case 3:
					d.hubClosePanel(now)
					d.showSheet(true)
					d.EditNewAlarm()
				case 4:
					d.hubClosePanel(now)
					d.showSheet(true)
				}
			} else if h.rings && p.In(d.r.hubPanelBox(true)) {
				if p.In(d.r.hubDotsBox(true)) {
					d.hubSetSlider(3, g.X)
				}
			} else if p.In(box) {
				dots := d.r.hubDotsBox(false)
				switch {
				case p.In(dots):
					row := 2
					if h.panel == hubBrightness {
						row = 1
					}
					d.hubSetSlider(row, g.X)
				case g.X < dots.Min.X:
					if h.panel == hubBrightness {
						d.setAuto(!d.autoValue(), true)
					} else {
						d.mu.Lock()
						d.hub.panel = hubMenu
						d.hub.rings = false
						d.mu.Unlock()
					}
				case g.X >= dots.Max.X && h.panel == hubVolume:
					d.mu.Lock()
					if d.hub.rings {
						d.hub.ringClosing = now
					} else {
						d.hub.rings = true
						d.hub.ringAt = now
						d.hub.ringClosing = time.Time{}
					}
					d.mu.Unlock()
				}
			} else {
				d.hubClosePanel(now)
			}
		case touch.Hold:
			row := 0
			p := image.Pt(g.X, g.Y)
			switch {
			case h.rings && p.In(d.r.hubDotsBox(true)):
				row = 3
			case h.panel == hubBrightness && p.In(d.r.hubDotsBox(false)):
				row = 1
			case h.panel == hubVolume && p.In(d.r.hubDotsBox(false)):
				row = 2
			}
			d.mu.Lock()
			d.hub.slider = row
			d.mu.Unlock()
			if row != 0 {
				d.hubSetSlider(row, g.X)
			}
		case touch.Drag, touch.Release:
			if h.slider != 0 {
				d.hubSetSlider(h.slider, g.X)
			}
			if g.Kind == touch.Release {
				d.mu.Lock()
				d.hub.slider = 0
				d.mu.Unlock()
			}
		}
		d.wake()
		return true
	}
	if h.smart {
		if !d.dashboardVisible() {
			if g.Kind == touch.SwipeRight || g.Kind == touch.SwipeUp || (g.Kind == touch.Tap && g.Y < d.r.h/5) {
				d.hubCloseSmart(now)
			}
			if g.Kind == touch.Tap && g.Y > d.r.h/2 {
				d.showSheet(true)
			}
			return true
		}
		// Native/streamed dashboard drags continue through its own hit testing; its normal left-edge
		// return calls closeDashboard, which also starts our return animation.
		if g.Kind == touch.SwipeRight {
			d.closeDashboard()
			return true
		}
		return false
	}
	if !d.onClock() || d.answerUp(now) {
		return false
	}
	edge := max(d.r.s(28), d.r.w/25)
	switch g.Kind {
	case touch.SwipeUp:
		if g.Y >= d.r.h-max(d.r.s(32), d.r.h/15) {
			d.mu.Lock()
			d.hub.panel = hubMenu
			d.hub.openingContact = contact
			d.hub.opened = now
			d.hub.closing = time.Time{}
			d.hub.touched = now
			d.hub.rings = false
			d.mu.Unlock()
			d.wake()
		}
		return true
	case touch.SwipeRight:
		if g.X < edge {
			d.mu.Lock()
			d.hub.night = true
			d.hub.page(image.Pt(1, 0), now)
			d.mu.Unlock()
			d.relight(true)
			d.wake()
		}
		return true
	case touch.SwipeLeft:
		if g.X >= d.r.w-edge {
			d.hubOpenSmart(image.Pt(-1, 0), now)
		}
		return true
	case touch.SwipeDown:
		if g.Y < max(d.r.s(28), d.r.h/15) {
			d.hubOpenSmart(image.Pt(0, 1), now)
		}
		return true
	}
	return false
}

func (d *Display) autoValue() bool        { d.mu.Lock(); defer d.mu.Unlock(); return d.autoOn }
func (d *Display) dashboardVisible() bool { d.mu.Lock(); defer d.mu.Unlock(); return d.dashShowing }
func (d *Display) hubOpenSmart(dir image.Point, now time.Time) {
	if dashboard.Get().Mode() != config.DashboardOff {
		d.openDashboard()
	}
	d.mu.Lock()
	d.hub.smart = true
	d.hub.openingContact, _ = touch.Get().Contact()
	d.hub.page(dir, now)
	d.mu.Unlock()
	d.wake()
}
func (d *Display) hubCloseSmart(now time.Time) {
	d.mu.Lock()
	d.hub.smart = false
	d.hub.page(image.Pt(1, 0), now)
	d.mu.Unlock()
	d.wake()
}

// The slider's labels are 1..10 for light and 0..10 for sound. Persist only a changed step, rather
// than every raw motion sample. Changing light manually also exits auto mode.
func (d *Display) hubSetSlider(row, x int) {
	n := d.r.hubDotStep(x, row == 3)
	switch row {
	case 1:
		n = max(1, n)
		if d.autoValue() {
			d.setAuto(false, true)
		}
		if d.ceilingOrDefault() != n*10 {
			d.apply(true, n*10, true)
		}
	case 2:
		n = hubFromTen(n, media.VolumeSteps)
		if media.Get().Volume() != n {
			media.Get().Set(n)
		}
	case 3:
		n = hubFromTen(n, media.VolumeSteps)
		if ring.Level() != n {
			alarm.Get().SetRingVolume(n, false)
		}
	}
}

// Use the existing hardware microphone switch and its persistent state.
var flipHubMicrophone = func() { mute.Get().Flip() }

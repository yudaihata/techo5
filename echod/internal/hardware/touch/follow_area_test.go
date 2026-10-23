//go:build !dot

package touch

import (
	"image"
	"testing"
	"time"
)

func TestLocalSliderFollowsOnlyItsOwnContact(t *testing.T) {
	for _, inside := range []bool{false, true} {
		s := &Screen{rawW: 800, rawH: 1280}
		f := &finger{sx: 400, sy: 500, x: 450, y: 500, at: time.Now()}
		x, y := s.landscape(f.sx, f.sy)
		area := image.Rect(x-5, y-5, x+5, y+5)
		if !inside {
			area = area.Add(image.Pt(200, 200))
		}
		s.SetFollowArea(area)
		var got []Gesture
		s.Gestures.Listen(func(g Gesture) { got = append(got, g) })
		s.moved(f)
		if inside && (len(got) < 2 || got[0].Kind != Hold || got[1].Kind != Drag) {
			t.Fatal("local slider did not follow", got)
		}
		if !inside {
			for _, g := range got {
				if g.Kind == Hold || g.Kind == Drag {
					t.Fatal("slider captured outside contact", got)
				}
			}
		}
	}
}

func TestOpeningSwipeDoesNotBecomeSliderDrag(t *testing.T) {
	s := &Screen{rawW: 800, rawH: 1280}
	f := &finger{sx: 400, sy: 500, x: 450, y: 500, swiped: true, at: time.Now()}
	s.SetFollowArea(image.Rect(0, 0, Width, Height))
	s.Gestures.Listen(func(g Gesture) {
		if g.Kind == Hold || g.Kind == Drag {
			t.Fatal("opening finger became a slider drag")
		}
	})
	s.moved(f)
}

func TestContactIdentityChangesOnlyOnNewFinger(t *testing.T) {
	s := &Screen{}
	s.setDown(true)
	first, down := s.Contact()
	if !down || first == 0 {
		t.Fatal("missing contact")
	}
	s.setDown(true)
	again, _ := s.Contact()
	if first != again {
		t.Fatal("same finger changed identity")
	}
	s.setDown(false)
	_, down = s.Contact()
	if down {
		t.Fatal("lift not reported")
	}
	s.setDown(true)
	next, _ := s.Contact()
	if next == first {
		t.Fatal("new finger shares opening swipe identity")
	}
}

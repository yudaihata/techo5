//go:build !dot && !spot

package display

import (
	"bytes"
	"embed"
	"image"
	"image/draw"
	"image/png"
	"sync"

	xdraw "golang.org/x/image/draw"
)

// Static Flat artwork, including its MIT copyright notice, travels with the daemon.
//
//go:embed assets/meteocons-flat/*.png assets/meteocons-flat/LICENSE
var photoWeatherAssets embed.FS

var photoWeatherCache sync.Map

type photoWeatherKey struct {
	name string
	size int
}

func photoWeatherName(condition string) string {
	switch condition {
	case "sunny":
		return "clear-day"
	case "clear-night":
		return "clear-night"
	case "cloudy":
		return "cloudy"
	case "partlycloudy":
		return "partly-cloudy-day"
	case "partlycloudy-night":
		return "partly-cloudy-night"
	case "rainy":
		return "rain"
	case "pouring":
		return "extreme-rain"
	case "lightning":
		return "thunderstorms"
	case "lightning-rainy":
		return "thunderstorms-rain"
	case "snowy":
		return "snow"
	case "snowy-rainy":
		return "sleet"
	case "hail":
		return "hail"
	case "fog":
		return "fog"
	case "windy", "windy-variant":
		return "wind"
	case "exceptional":
		return "hurricane"
	default:
		return "not-available"
	}
}

// Keep decoded and scaled images across clock ticks; no SVG engine or network at runtime.
func photoWeatherImage(condition string, size int) image.Image {
	name := photoWeatherName(condition)
	key := photoWeatherKey{name, size}
	if cached, ok := photoWeatherCache.Load(key); ok {
		return cached.(image.Image)
	}
	originalKey := photoWeatherKey{name, 0}
	original, ok := photoWeatherCache.Load(originalKey)
	if !ok {
		b, err := photoWeatherAssets.ReadFile("assets/meteocons-flat/" + name + ".png")
		if err != nil {
			return nil
		}
		img, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			return nil
		}
		original, _ = photoWeatherCache.LoadOrStore(originalKey, img)
	}
	out := image.NewRGBA(image.Rect(0, 0, size, size))
	src := original.(image.Image)
	xdraw.CatmullRom.Scale(out, out.Rect, src, src.Bounds(), draw.Src, nil)
	cached, _ := photoWeatherCache.LoadOrStore(key, image.Image(out))
	return cached.(image.Image)
}

func (r *renderer) photoWeatherIcon(condition string, box image.Rectangle) {
	if icon := photoWeatherImage(condition, box.Dx()); icon != nil {
		draw.Draw(r.dst, box, icon, icon.Bounds().Min, draw.Over)
	}
}

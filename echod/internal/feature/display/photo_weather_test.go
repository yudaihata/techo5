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
)

func TestPhotoWeatherAssets(t *testing.T) {
	conditions := []string{"sunny", "clear-night", "cloudy", "partlycloudy", "partlycloudy-night", "rainy", "pouring", "lightning", "lightning-rainy", "snowy", "snowy-rainy", "hail", "fog", "windy", "windy-variant", "exceptional", "future-condition"}
	for _, condition := range conditions {
		for _, size := range []int{36, 48, 72} {
			icon := photoWeatherImage(condition, size)
			if icon == nil || icon.Bounds().Dx() != size || icon.Bounds().Dy() != size {
				t.Fatalf("%s/%d: missing or incorrect asset", condition, size)
			}
			if icon != photoWeatherImage(condition, size) {
				t.Fatalf("%s/%d: icon was not cached", condition, size)
			}
			opaque, transparent := 0, 0
			for y := 0; y < size; y++ {
				for x := 0; x < size; x++ {
					_, _, _, a := icon.At(x, y).RGBA()
					if a == 0 {
						transparent++
					} else {
						opaque++
					}
				}
			}
			if opaque < size || transparent < size {
				t.Errorf("%s/%d: empty icon or lost transparency", condition, size)
			}
		}
	}
	if photoWeatherName("partlycloudy") == photoWeatherName("partlycloudy-night") {
		t.Fatal("day and night share an icon")
	}
	if photoWeatherName("rainy") == photoWeatherName("pouring") {
		t.Fatal("rain severity is lost")
	}
}

func TestPhotoWeatherPreview(t *testing.T) {
	conditions := []string{"sunny", "clear-night", "cloudy", "partlycloudy", "partlycloudy-night", "rainy", "pouring", "lightning", "lightning-rainy", "snowy", "snowy-rainy", "hail", "fog", "windy", "exceptional", "unknown"}
	canvas := image.NewRGBA(image.Rect(0, 0, 768, 448))
	r := newRenderer(canvas)
	for i, condition := range conditions {
		x, y := i%4*192, i/4*112
		bg := color.RGBA{30, 38, 50, 255}
		ink := color.RGBA{255, 255, 255, 255}
		if i%4 >= 2 {
			bg = color.RGBA{240, 242, 245, 255}
			ink = color.RGBA{30, 38, 50, 255}
		}
		draw.Draw(canvas, image.Rect(x, y, x+192, y+112), image.NewUniform(bg), image.Point{}, draw.Src)
		r.photoWeatherIcon(condition, image.Rect(x+64, y+4, x+128, y+68))
		r.text(r.micro, condition, x+10, y+96, ink)
	}
	if dir := os.Getenv("SHOW_PREVIEW"); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(filepath.Join(dir, "photo-weather-flat.png"))
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

package fonts

import (
	"bytes"
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
)

func TestNotoPrimary(t *testing.T) {
	if !Known(NotoSansCJKJP) || Names()[Index(NotoSansCJKJP)] != "Noto Sans CJK JP" {
		t.Fatal("Noto is not a primary font choice")
	}
	for _, bold := range []bool{false, true} {
		f, err := Font(NotoSansCJKJP, bold)
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range "0123456789ABCxyz日本語漢字廊下の時計天気曇雨雪あいうアイウ。、（）" {
			if g, err := f.GlyphIndex(nil, r); err != nil || g == 0 {
				t.Errorf("bold=%v: missing %q", bold, r)
			}
		}
		family, err := f.Name(nil, sfnt.NameIDFamily)
		if err != nil || family != "Noto Sans CJK JP" {
			t.Fatalf("wrong family %q: %v", family, err)
		}
		face, err := Face(NotoSansCJKJP, bold, 40)
		if err != nil {
			t.Fatal(err)
		}
		expected, err := opentype.NewFace(f, &opentype.FaceOptions{Size: 40 * families[Index(NotoSansCJKJP)].scale, DPI: 72, Hinting: font.HintingFull})
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range "8A日語" {
			a, ok := face.GlyphAdvance(r)
			want, _ := expected.GlyphAdvance(r)
			if !ok || a != want {
				t.Errorf("%q uses fallback instead of Noto", r)
			}
		}
		for _, r := range "ŸŽŁŚŹŻĆĄĘČŘŠŤŮ" {
			if stack, ok := face.(*Stack); !ok || !stack.Has(r) {
				t.Errorf("extended Latin fallback missing %q", r)
			}
		}
		face.Close()
		expected.Close()
	}
	regular, _ := files.ReadFile("files/notosanscjkjp/Regular.otf")
	bold, _ := files.ReadFile("files/notosanscjkjp/Bold.otf")
	if bytes.Equal(regular, bold) {
		t.Fatal("Bold is not a distinct cut")
	}
}

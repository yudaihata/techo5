//go:build !dot

package display

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/HuskerMinion/techo5/echod/internal/config"
	"github.com/HuskerMinion/techo5/echod/internal/lib/fonts"
)

func TestNotoDeviceFontSelection(t *testing.T) {
	config.Use(filepath.Join(t.TempDir(), "state.json"))
	d := Get()
	if !slices.Contains(d.fontSel.Options, "Noto Sans CJK JP") {
		t.Fatal("actual device font select does not offer Noto")
	}
	d.fontSel.OnCommand("Noto Sans CJK JP")
	if config.Get().Screen.Font != fonts.NotoSansCJKJP || screenFont() != fonts.NotoSansCJKJP || fontRow().value != "Noto Sans CJK JP" {
		t.Fatal("the device did not save and show Noto")
	}
	d.fontSel.OnCommand("Go")
	if screenFont() != fonts.Go {
		t.Fatal("switching back to Go failed")
	}
}

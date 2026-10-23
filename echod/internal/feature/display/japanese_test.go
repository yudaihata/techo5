//go:build !dot

package display

import (
	"path/filepath"
	"testing"

	"github.com/HuskerMinion/techo5/echod/internal/config"
)

func TestJapaneseScreenLanguageSelection(t *testing.T) {
	config.Use(filepath.Join(t.TempDir(), "state.json"))
	t.Cleanup(func() { _ = config.Set().Screen().Language("") })
	selectLanguage := langSelect()
	selectLanguage.OnCommand("日本語")
	if got := config.Get().Screen.Language; got != "ja" {
		t.Fatalf("Japanese selection saved %q", got)
	}
	if langCodes[langIndex()] != "ja" || langOptions[langIndex()] != "日本語" || screenLang() != "ja" {
		t.Fatal("the settings row and Home Assistant disagree about Japanese")
	}
	if conditionWords("partlycloudy") != "晴れ時々曇り" || colorWord(2) != "完了" || climateWord("heat") != "暖房" || mediaWord(6) != "再生中" {
		t.Fatal("selecting Japanese did not translate the screen's words")
	}
	selectLanguage.OnCommand("English")
	if screenLang() != "en" || conditionWords("partlycloudy") != "Partly cloudy" {
		t.Fatal("switching back to English did not restore English")
	}
}

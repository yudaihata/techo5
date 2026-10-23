//go:build !dot && !spot

package display

import (
	"github.com/HuskerMinion/techo5/echod/internal/config"
	"path/filepath"
	"testing"
)

func TestJapaneseSheetWords(t *testing.T) {
	config.Use(filepath.Join(t.TempDir(), "state.json"))
	if err := config.Set().Screen().Language("ja"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.Set().Screen().Language("") })
	if colorWord(2) != "完了" || climateWord("heat") != "暖房" || mediaWord(6) != "再生中" {
		t.Fatal("the Japanese sheets did not use their translations")
	}
}

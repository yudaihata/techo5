//go:build !dot && !spot

package display

func init() {
	colorWords["ja"] = [4]string{"暖色", "寒色", "完了", "グループ · 調色非対応のライトは白色のままです"}
	climateWords["ja"] = map[string]string{
		"now": "現在", "low": "暖房設定", "high": "冷房設定", "done": "完了",
		"off": "オフ", "heat": "暖房", "cool": "冷房", "heat_cool": "冷暖房", "auto": "自動", "dry": "除湿", "fan_only": "送風",
	}
	mediaWords["ja"] = []string{"再生", "スピーカー", "完了", "読み込み中…", "Music Assistantにお気に入りはありません",
		"この機器ではMusic Assistantの操作のみ可能です", "再生中", "一時停止", "待機中", "オフ",
		"お気に入りにはHome Assistant管理者のトークンが必要です"}
}

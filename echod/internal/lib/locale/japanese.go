package locale

// Keep the additional language together so upstream's other translations can evolve independently.
func init() {
	langs["ja"] = &words{
		days:        [7]string{"日曜日", "月曜日", "火曜日", "水曜日", "木曜日", "金曜日", "土曜日"},
		shortDays:   [7]string{"日", "月", "火", "水", "木", "金", "土"},
		months:      [12]string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"},
		shortMonths: [12]string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"},
		sky: map[string]string{
			"clear-night": "快晴", "cloudy": "曇り", "exceptional": "荒天", "fog": "霧", "hail": "ひょう",
			"lightning": "雷", "lightning-rainy": "雷雨", "partlycloudy": "晴れ時々曇り",
			"pouring": "大雨", "rainy": "雨", "snowy": "雪", "snowy-rainy": "みぞれ",
			"sunny": "晴れ", "windy": "強風", "windy-variant": "強風",
		},
	}
	phrases["ja"] = map[string]string{
		"here": "現在地", "today": "今日", "alarm": "アラーム", "snoozed": "スヌーズ",
		"tomorrow": "明日", "yesterday": "昨日", "now": "現在", "allday": "終日",
		"nothing": "予定はありません", "nothingtoday": "今日の予定はありません",
		"nothingelse": "今日の残りの予定はありません", "upcoming": "今後の予定", "nothingsoon": "直近の予定はありません",
	}
	updateTranslations["ja"] = map[string]string{
		UpdateSystemTitle: "システム更新", UpdateFailedTitle: "更新に失敗しました",
		UpdateDownloading: "ダウンロード中 · %d%%", UpdateInstalling: "インストール中…",
		UpdateRestartingIn: "%d秒後にアプリを再起動…", UpdateRestarting: "アプリを再起動中…",
		UpdateRebootingIn: "%d秒後に本体を再起動…", UpdateRebooting: "本体を再起動中…",
		UpdatePowerHint: "電源を切らないでください", UpdateRetryHint: "Settings > General > Updatesから再試行してください",
	}
}

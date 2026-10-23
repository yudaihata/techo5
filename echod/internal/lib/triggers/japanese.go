package triggers

func init() {
	Languages["ja"] = Set{
		// Avoid single characters such as 雨 (also in 雨音) and 雪 (also a person's name).
		Weather:    []string{"天気", "気温", "降水", "雨が", "雨は", "雨降", "雨降り", "雨雲", "雪が", "雪は", "雪降", "雪かな", "雪の予報", "台風", "雷雨", "レーダー", "雨の地図"},
		Radar:      []string{"雨雲", "降水レーダー", "天気図", "雨の地図"},
		NotWeather: []string{"雨音", "雨の音", "雨が降る音", "雨のサウンド", "雨の曲", "再生", "流して"},
		GoHome:     []string{"ホーム画面", "メイン画面", "時計に戻", "時計を表示", "時計を見せ"},
		Camera:     []string{"カメラ", "映して", "見せて", "表示して"},
	}
}

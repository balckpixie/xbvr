package shared

import (
	"fmt"
	"strings"
)

// ひらがな → ローマ字（ヘボン式）変換マップ（基本形）
var hiraToRomaji = map[string]string{
	"あ": "a", "い": "i", "う": "u", "え": "e", "お": "o",
	"か": "ka", "き": "ki", "く": "ku", "け": "ke", "こ": "ko",
	"さ": "sa", "し": "shi", "す": "su", "せ": "se", "そ": "so",
	"た": "ta", "ち": "chi", "つ": "tsu", "て": "te", "と": "to",
	"な": "na", "に": "ni", "ぬ": "nu", "ね": "ne", "の": "no",
	"は": "ha", "ひ": "hi", "ふ": "fu", "へ": "he", "ほ": "ho",
	"ま": "ma", "み": "mi", "む": "mu", "め": "me", "も": "mo",
	"や": "ya", "ゆ": "yu", "よ": "yo",
	"ら": "ra", "り": "ri", "る": "ru", "れ": "re", "ろ": "ro",
	"わ": "wa", "を": "wo", "ん": "n",
	"が": "ga", "ぎ": "gi", "ぐ": "gu", "げ": "ge", "ご": "go",
	"ざ": "za", "じ": "ji", "ず": "zu", "ぜ": "ze", "ぞ": "zo",
	"だ": "da", "ぢ": "ji", "づ": "zu", "で": "de", "ど": "do",
	"ば": "ba", "び": "bi", "ぶ": "bu", "べ": "be", "ぼ": "bo",
	"ぱ": "pa", "ぴ": "pi", "ぷ": "pu", "ぺ": "pe", "ぽ": "po",
	"きゃ": "kya", "きゅ": "kyu", "きょ": "kyo",
	"しゃ": "sha", "しゅ": "shu", "しょ": "sho",
	"ちゃ": "cha", "ちゅ": "chu", "ちょ": "cho",
	"にゃ": "nya", "にゅ": "nyu", "にょ": "nyo",
	"ひゃ": "hya", "ひゅ": "hyu", "ひょ": "hyo",
	"みゃ": "mya", "みゅ": "myu", "みょ": "myo",
	"りゃ": "rya", "りゅ": "ryu", "りょ": "ryo",
	"ぎゃ": "gya", "ぎゅ": "gyu", "ぎょ": "gyo",
	"じゃ": "ja", "じゅ": "ju", "じょ": "jo",
	"びゃ": "bya", "びゅ": "byu", "びょ": "byo",
	"ぴゃ": "pya", "ぴゅ": "pyu", "ぴょ": "pyo",
	"っ": "", // 小さい「っ」は後続の子音を繰り返す（後で処理）
}

// ひらがなをローマ字に変換する関数
func HiraganaToRomaji(input string) string {
	var result strings.Builder

	runes := []rune(input)
	for i := 0; i < len(runes); {
		// 2文字（例: きゃ, しゃ）の合成を優先
		if i+1 < len(runes) {
			pair := string(runes[i]) + string(runes[i+1])
			if romaji, ok := hiraToRomaji[pair]; ok {
				result.WriteString(romaji)
				i += 2
				continue
			}
		}

		char := string(runes[i])

		// 小さい「っ」の処理（次の文字の先頭子音を繰り返す）
		if char == "っ" && i+1 < len(runes) {
			nextChar := string(runes[i+1])
			nextRomaji := ""
			if i+2 < len(runes) {
				pair := nextChar + string(runes[i+2])
				if r, ok := hiraToRomaji[pair]; ok {
					nextRomaji = r
				}
			}
			if nextRomaji == "" {
				if r, ok := hiraToRomaji[nextChar]; ok && len(r) > 0 {
					result.WriteByte(r[0]) // 先頭の子音を繰り返す
				}
			}
			i++
			continue
		}

		// 通常の1文字処理
		if romaji, ok := hiraToRomaji[char]; ok {
			result.WriteString(romaji)
		} else {
			result.WriteString(char) // 未知の文字はそのまま出力
		}
		i++
	}
	return result.String()
}

func main() {
	sample := "きょうはいいてんきですね"
	fmt.Println(HiraganaToRomaji(sample)) // kyouhaiitenkidesune
}

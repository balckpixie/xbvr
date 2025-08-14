package shared

import (
	"encoding/json"
	"regexp"
	"strings"
)

func FilterHiraganaOnly(inputText string) []string {
	// 入力のJSON文字列を []string に変換
	var words []string
	if err := json.Unmarshal([]byte(inputText), &words); err != nil {
		return nil
	}

	// 正規表現: ひらがな＋長音記号・繰り返し記号のみ
	hiraganaRegex := regexp.MustCompile(`^[\x{3041}-\x{3096}ーゝゞ\s（）]+$`)

	var result []string
	for _, w := range words {
		trimmed := strings.TrimSpace(w)
		if hiraganaRegex.MatchString(trimmed) {
			result = append(result, trimmed)
		}
	}

	return result
}


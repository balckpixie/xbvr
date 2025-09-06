package shared

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

func IsQuoted(input string) bool {
	// 文字列がダブルクオーテーションで始まり、終わるかどうかを確認
	return strings.HasPrefix(input, "\"") && strings.HasSuffix(input, "\"")
}

func GetQuotedString(input string) (string, error) {
	// ダブルクオーテーションで始まり、終わるかどうかを確認
	if strings.HasPrefix(input, "\"") && strings.HasSuffix(input, "\"") {
		// ダブルクオーテーションを除去して返す
		return strings.Trim(input, "\""), nil
	}
	// ダブルクオーテーションで囲まれた文字列が見つからない場合、エラーを返す
	return "", errors.New("quoted string not found")
}

func GetFirstCharsFromJSON(jsonStr string, count int) (string, error) {
	var arr []string
	err := json.Unmarshal([]byte(jsonStr), &arr)
	if err == nil {
		// JSON配列として処理
		if len(arr) > 0 && len(arr[0]) > 0 {
			return truncateUTF8String(arr[0], count), nil
		}
		return "", fmt.Errorf("empty array or empty first element")
	}

	// JSONでない場合は普通の文字列として処理
	if utf8.RuneCountInString(jsonStr) > 0 {
		return truncateUTF8String(jsonStr, count), nil
	}

	return "", fmt.Errorf("empty string")
}

// UTF-8エンコードを考慮して文字列を切り詰める
func truncateUTF8String(s string, count int) string {
    runeCount := 0
    for i := range s {
        runeCount++
        if runeCount > count {
            return s[:i]
        }
    }
    return s
}
package shared

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)


// DVD-ID形式をDMM形式に変換
func ConvertFormat(input string) string {
	re := regexp.MustCompile(`([a-zA-Z]{3,4})-(\d{3,})`)
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		prefix := match[1]
		numStr := match[2]

		// 数字が4桁以上の場合はゼロパディングを行わない
		num, err := strconv.Atoi(numStr)
		if err == nil {
			if num < 10000 {
				numStr = fmt.Sprintf("%05d", num)
			}
		}
		replacement := prefix + numStr
		input = strings.Replace(input, match[0], replacement, -1)
	}

	return input
}

// DMM形式をDVD-ID形式に変換
func ConvertToDVDId(input string) string {
	re := regexp.MustCompile(`^([H_]*)(\d*)([a-zA-Z]+)(\d+)`)
	matches := re.FindStringSubmatch(input)

	if len(matches) == 5 {
		prefix := strings.ToUpper(strings.Replace(matches[3], "_", "", 1))
		num, err := strconv.Atoi(matches[4])
		if err == nil {
			numStr := fmt.Sprintf("%03d", num) // 3桁の数字に変換
			return fmt.Sprintf("%s-%s", prefix, numStr)
		}
	}
	return ""
}

func ExtractDVDIDLogic(filename string) string {
	dvdid := ""
	regex := regexp.MustCompile(`[a-zA-Z0-9]{2,6}-\d{2,6}`)
	match := regex.FindString(filename)
	if match == "" {
		regex = regexp.MustCompile(`([a-zA-Z]{2,6})(\d{2,6})`)
		match := regex.FindStringSubmatch(filename)
		if match != nil {
			firstPart := match[1]
			secondPart := match[2]
			if len(secondPart) >= 4 {
				secondPart = strings.TrimLeft(secondPart, "0")
			}
			// 確認: secondPart が 3 文字未満の場合は 0 を補完する
			if len(secondPart) < 3 {
				secondPart = strings.Repeat("0", 3-len(secondPart)) + secondPart
			}
			dvdid = firstPart + "-" + secondPart
		} else {
			dvdid = "null"
		}
	} else {
		dvdid = match
	}
	return dvdid
}

func ExtractFormat(input string) []string {
	re := regexp.MustCompile(`[A-Z]{4}-\d{3,4}`)
	return re.FindAllString(input, -1)
}
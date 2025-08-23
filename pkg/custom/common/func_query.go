package common

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/xbapps/xbvr/pkg/config"
)

// URLとドメインの一致を確認
func DomainMatch(urlString, pattern string) bool {
	u, err := url.Parse(urlString)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return false
	}
	host := u.Hostname()

	pattern = "^" + regexp.QuoteMeta(pattern) + "$"
	pattern = strings.ReplaceAll(pattern, "\\*", "[^.]+") // ワイルドカードを正規表現に変換
	re := regexp.MustCompile(pattern)
	isMatch := re.MatchString(host)

	return isMatch
}

// DMM API呼び出し時のQueryパラメータを追加
func AddAPIParam(originalURL string) (string, error) {

	if config.Config.Custom.DmmAPIKey.DmmApiId == "" || config.Config.Custom.DmmAPIKey.DmmAffiliateId == "" {
		return "", errors.New("is not set DmmApiId and DmmAffiliateId param")
	}

	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return "", err
	}

	queryParams := parsedURL.Query()
	queryParams.Set("api_id", config.Config.Custom.DmmAPIKey.DmmApiId)
	queryParams.Set("affiliate_id", config.Config.Custom.DmmAPIKey.DmmAffiliateId)
	queryParams.Set("offset", "1")
	parsedURL.RawQuery = queryParams.Encode()

	return parsedURL.String(), nil
}

// Query Parameterを追加する
func AddQueryParam(originalURL string, paramname string, value string) (string, error) {
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return "", err
	}
	queryParams := parsedURL.Query()
	queryParams.Set(paramname, value)
	parsedURL.RawQuery = queryParams.Encode()
	return parsedURL.String(), nil
}

// Query Parameterの名前を置き換える
func ReplaceQueryParam(originalURL string, paramname string, newParamname string) (string, error) {
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return "", err
	}

	queryParams := parsedURL.Query()
	values := queryParams[paramname]
	if len(values) > 0 {
		// パラメーターが見つかった場合、新しいパラメーター名に置き換える
		queryParams.Del(paramname)
		for _, value := range values {
			queryParams.Add(newParamname, value)
		}
		parsedURL.RawQuery = queryParams.Encode()
	} else {
		return "", errors.New("not found parameter")
	}

	return parsedURL.String(), nil
}
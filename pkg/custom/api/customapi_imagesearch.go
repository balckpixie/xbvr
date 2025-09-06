package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/xbapps/xbvr/pkg/models"

	"github.com/gocolly/colly/v2"
)

type ImageItem struct {
	ID  string `json:"id"`
	Src string `json:"src"`
	Alt string `json:"alt"`
}

type ResponseGetImages struct {
	Results int      `json:"results"`
	Images  []string `json:"images"`
}

type SearchImageResponse struct {
	Results int         `json:"results"`
	Images  []ImageItem `json:"images"`
}

type RequestSearchImages struct {
	ActorID uint     `json:"actor_id"`
	Url     string   `json:"url"`
	Keyword string   `json:"keyword"`
	Site    SiteType `json:"site"`
}

type ImagesResource struct{}

type SiteType string

const (
	SiteGoogle SiteType = "Google"
	SiteBing   SiteType = "Bing"
)

type SiteConfig struct {
	BaseURL      string
	QueryPattern string // %s に検索キーワードを埋め込む
}

// 検索サイト設定マップ
var SiteConfigs = map[SiteType]SiteConfig{
	SiteGoogle: { // Google画像検索
		BaseURL:      "https://www.google.com/search",
		QueryPattern: "q=%s&as_epq=&as_oq=&as_eq=&imgar=t|xt&imgcolor=&imgtype=photo&cr=&as_sitesearch=&as_filetype=&tbs=&udm=2&no_sw_cr=1&sssc=1",
	},
	SiteBing: { // Bing画像検索
		BaseURL:      "https://www.bing.com/images/search",
		QueryPattern: "q=%s&qft=+filterui:aspect-tall",
	},
}

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36"
//const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99 Safari/537.36"

func (i ImagesResource) WebService() *restful.WebService {
	tags := []string{"CustomImages"}

	ws := new(restful.WebService)

	ws.Path("/api_custom/images").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/searchImage").To(i.searchActorImage).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(SearchImageResponse{}))

	return ws
}

func (i ImagesResource) searchActorImage(req *restful.Request, resp *restful.Response) {
	var r RequestSearchImages
	err := req.ReadEntity(&r)
	if err != nil {
		log.Error(err)
		return
	}

	if r.ActorID == 0 {
		return
	}

	db, _ := models.GetDB()
	defer db.Close()

	var actor models.Actor
	err = actor.GetIfExistByPKWithSceneAvg(r.ActorID)
	if err != nil {
		log.Error(err)
		return
	}

	var imageURLs []string
	query := "(\"" + actor.Name + "\") " + r.Keyword
	switch r.Site {
	case SiteGoogle:
		imageURLs, err = getImageURLsFromGoogle(query)
	case SiteBing:
		imageURLs, err = getImageURLsFromBing(query)
	default:
		log.Errorf("invalid site value: %s", r.Site)
		resp.WriteErrorString(http.StatusBadRequest, "Invalid site parameter")
		return
	}

	if err != nil {
		log.Error(err)
		return
	}

	if len(imageURLs) == 0 {
		resp.WriteHeaderAndEntity(http.StatusOK, SearchImageResponse{
			Results: 0,
			Images:  []ImageItem{},
		})
		return
	}

	var images []ImageItem
	for idx, url := range imageURLs {
		images = append(images, ImageItem{
			ID:  fmt.Sprintf("%d", idx+1),
			Src: url,
			Alt: "",
		})
	}

	resp.WriteHeaderAndEntity(http.StatusOK, SearchImageResponse{
		Results: len(images),
		Images:  images,
	})
}

func getImageURLsFromGoogle(query string) ([]string, error) {
	// Cookie取得
	// cookies, err := fetchGoogleCookies()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to fetch cookies: %w", err)
	// }
	searchURL, err := buildSearchURL(SiteGoogle, url.QueryEscape(query))
	if err != nil {
		return nil, err
	}

	fmt.Println("Google Search URL:", searchURL)
	
	imageCollector := colly.NewCollector(
		colly.UserAgent(userAgent),
	)
	imageCollector.OnRequest(func(r *colly.Request) {
		// imageCollector.SetCookies("http://www.google.com", cookies)
	})

	var imageURLs []string
	var parseErr error

	imageCollector.OnHTML("script", func(e *colly.HTMLElement) {
		// contentType := e.Response.Headers.Get("Content-Type")
		// fmt.Println("Content-Type:", contentType)

		googleScriptRe := regexp.MustCompile(`(?s)var m\s*=\s*(\{.*?\});`)

		matches := googleScriptRe.FindStringSubmatch(e.Text)
		if len(matches) < 1 {
			return
		}
		var mObj map[string]interface{}
		if err := json.Unmarshal([]byte(matches[1]), &mObj); err != nil {
			parseErr = fmt.Errorf("failed to parse JSON: %w", err)
			return
		}

		imageURLs = append(imageURLs, extractGoogleImageURLs(mObj)...)
	})

	imageCollector.OnError(func(r *colly.Response, err error) {
		parseErr = fmt.Errorf("google request failed: %w", err)
	})

	if err := imageCollector.Visit(searchURL); err != nil {
		return nil, fmt.Errorf("failed to visit Google search URL: %w", err)
	}

	if parseErr != nil {
		return nil, parseErr
	}

	return imageURLs, nil
}

func getImageURLsFromBing(query string) ([]string, error) {
	// Cookie取得
	cookies, err := fetchBingCookies()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cookies: %w", err)
	}

	searchURL, err := buildSearchURL(SiteBing, url.QueryEscape(query))
	if err != nil {
		return nil, err
	}

	var imageURLs []string
	var parseErr error

	imageCollector := colly.NewCollector(
		colly.UserAgent(userAgent),
	)

	imageCollector.OnRequest(func(r *colly.Request) {
		imageCollector.SetCookies("http://www.bing.com", cookies)
	})

	imageCollector.OnHTML("div.imgpt > a.iusc", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		if imageURL := extractImageURLFromBing(href); imageURL != "" {
			imageURLs = append(imageURLs, imageURL)
		}
	})

	imageCollector.OnError(func(r *colly.Response, err error) {
		parseErr = fmt.Errorf("bing request failed: %w", err)
	})

	if err := imageCollector.Visit(searchURL); err != nil {
		return nil, fmt.Errorf("failed to visit search URL: %w", err)
	}

	if parseErr != nil {
		return nil, parseErr
	}

	return imageURLs, nil

}

func buildSearchURL(site SiteType, keyword string) (string, error) {
	config, ok := SiteConfigs[site]
	if !ok {
		return "", fmt.Errorf("unsupported site: %s", site)
	}
	return fmt.Sprintf("%s?%s", config.BaseURL, fmt.Sprintf(config.QueryPattern, keyword)), nil
}

// extractGoogleImageURLs JSONオブジェクトからGoogle画像URL抽出
func extractGoogleImageURLs(mObj map[string]interface{}) []string {
	var urls []string

	for _, value := range mObj {
		arr1, ok := value.([]interface{})
		if !ok || len(arr1) < 2 {
			continue
		}
		arr2, ok := arr1[1].([]interface{})
		if !ok || len(arr2) < 4 {
			continue
		}
		arr3, ok := arr2[3].([]interface{})
		if !ok || len(arr3) < 1 {
			continue
		}
		if imgURL, ok := arr3[0].(string); ok && imgURL != "" {
			urls = append(urls, imgURL)
		}
	}

	return urls
}

// extractImageURLFromBing Bing画像検索結果のhrefから画像URLを抽出
func extractImageURLFromBing(href string) string {
	bingURLRe := regexp.MustCompile(`mediaurl=([^&]+)`)
	matches := bingURLRe.FindStringSubmatch(href)
	if len(matches) > 1 {
		if decodedURL, err := url.QueryUnescape(matches[1]); err == nil {
			return decodedURL
		}
	}
	return ""
}

// Googleのトップページを訪問してCookieを取得する
func fetchGoogleCookies() ([]*http.Cookie, error) {
	var cookies []*http.Cookie

	cookieCollector := colly.NewCollector(colly.UserAgent(userAgent))
	cookieCollector.OnResponse(func(r *colly.Response) {
		cookies = cookieCollector.Cookies(r.Request.URL.String())
	})
	config, _ := SiteConfigs[SiteGoogle]
	if err := cookieCollector.Visit(config.BaseURL); err != nil {
		return nil, err
	}
	return cookies, nil
}

// Bingのトップページを訪問してCookieを取得する
func fetchBingCookies() ([]*http.Cookie, error) {
	var cookies []*http.Cookie

	cookieCollector := colly.NewCollector(colly.UserAgent(userAgent))
	cookieCollector.OnResponse(func(r *colly.Response) {
		cookies = cookieCollector.Cookies(r.Request.URL.String())
		for _, cookie := range cookies {
			if cookie.Name == "SRCHHPGUSR" {
				cookie.Value += "&ADLT=OFF"
			}
		}
	})
	config, _ := SiteConfigs[SiteBing]
	if err := cookieCollector.Visit(config.BaseURL); err != nil {
		return nil, err
	}
	return cookies, nil
}

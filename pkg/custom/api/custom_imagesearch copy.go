package api

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"net/url"
// 	"regexp"

// 	restfulspec "github.com/emicklei/go-restful-openapi/v2"
// 	"github.com/emicklei/go-restful/v3"
// 	"github.com/xbapps/xbvr/pkg/models"

// 	"github.com/gocolly/colly/v2"
// )

// type ImageItem struct {
//     ID  string `json:"id"`
//     Src string `json:"src"`
//     Alt string `json:"alt"`
// }

// type ResponseGetImages struct {
// 	Results int            `json:"results"`
// 	Images  []string 		`json:"images"`
// }

// type SearchImageResponse struct {
// 	Results int            `json:"results"`
//     Images []ImageItem `json:"images"`
// }

// type RequestSearchImages struct {
// 	ActorID uint   		`json:"actor_id"`
// 	Url     string 		`json:"url"`
// 	Keyword string 		`json:"keyword"`
// 	Site 	SiteType  	`json:"site"`
// }

// type ImagesResource struct{}

// type SiteType string
// const (
// 	SiteGoogle SiteType = "Google"
// 	SiteBing SiteType = "Bing"
// )

// func (i ImagesResource) WebService() *restful.WebService {
// 	tags := []string{"CustomImages"}

// 	ws := new(restful.WebService)

// 	ws.Path("/api_custom/images").
// 		Consumes(restful.MIME_JSON).
// 		Produces(restful.MIME_JSON)

// 	ws.Route(ws.POST("/searchImage").To(i.searchActorImage).
// 		Metadata(restfulspec.KeyOpenAPITags, tags).
// 		Writes(SearchImageResponse{}))

// 	return ws
// }

// func (i ImagesResource) searchActorImage(req *restful.Request, resp *restful.Response) {
// 	var r RequestSearchImages
// 	err := req.ReadEntity(&r)
// 	if err != nil {
// 		log.Error(err)
// 		return
// 	}

// 	if r.ActorID == 0 {
// 		return
// 	}

// 	db, _ := models.GetDB()
// 	defer db.Close()

// 	var actor models.Actor
// 	err = actor.GetIfExistByPKWithSceneAvg(r.ActorID)
// 	if err != nil {
// 		log.Error(err)
// 		return
// 	}

// 	var imageURLs []string
// 	query := "(\"" + actor.Name + "\") " + r.Keyword
// 	if r.Site == SiteGoogle {
// 		imageURLs, err = getImageURLsFromGoogleImage2(query)
// 	} else if r.Site == SiteBing {
// 		imageURLs, err = getImageURLs2(query)
// 	} else {
// 		log.Errorf("invalid site value: %s", r.Site)
// 		resp.WriteErrorString(http.StatusBadRequest, "Invalid site parameter")
// 		return
// 	}
	
// 	if err != nil {
//         log.Error(err)
// 		return
// 	}
	
// 	    if len(imageURLs) == 0 {
//         resp.WriteHeaderAndEntity(http.StatusOK, SearchImageResponse{
//             Results: 0,
//             Images:  []ImageItem{},
//         })
//         return
//     }

//     var images []ImageItem
//     for idx, url := range imageURLs {
//         images = append(images, ImageItem{
//             ID:  fmt.Sprintf("%d", idx+1),
//             Src: url,
//             Alt: "",
//         })
//     }

//     resp.WriteHeaderAndEntity(http.StatusOK, SearchImageResponse{
//         Results: len(images),
//         Images:  images,
//     })
// }

// func getImageURLsFromGoogleImage2(query string) ([]string, error) {
// 	var imageURLs []string
// 	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s&as_epq=&as_oq=&as_eq=&imgar=t|xt&imgcolor=&imgtype=photo&cr=&as_sitesearch=&as_filetype=&tbs=&udm=2", url.QueryEscape(query))
// 	fmt.Println("searchURL:", searchURL)
// 	c := colly.NewCollector()
// 	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99 Safari/537.36"
// 	c.OnHTML("script", func(e *colly.HTMLElement) {
// 		// スクリプトコードを取得
// 		scriptContent := e.Text
// 		re := regexp.MustCompile(`var m=({[^;]*})`)
// 		matches := re.FindStringSubmatch(scriptContent)
	
// 		if len(matches) >= 2 {
// 			// 変数mの値を取得
// 			mValue := matches[1]
	
// 			// fmt.Println("mValue:", mValue)
// 			// JSON形式の文字列をパースしてオブジェクトに変換
// 			var mObj map[string]interface{}
// 			if err := json.Unmarshal([]byte(mValue), &mObj); err != nil {
// 				fmt.Println("JSONの解析中にエラーが発生しました:", err)
// 				return
// 			}
	
// 			// 変数mの値を出力
// 			// fmt.Println("変数mの値:")
// 			for _, value := range mObj {
// 				// fmt.Printf("%s: %v\n", key, value)
// 				if array, isArray := value.([]interface{}); isArray {
// 					if len(array) >= 2 {
// 						if secondArray, isSecondArray := array[1].([]interface{}); isSecondArray {
// 							if len(secondArray) >= 4 {
// 								if thirdArray, isThirdArray := secondArray[3].([]interface{}); isThirdArray {
// 									if len(thirdArray) >= 3 {
// 										imageURL := thirdArray[0].(string)
// 										if imageURL != "" {
// 											imageURLs = append(imageURLs, imageURL)
// 										}
// 									}
// 								}
// 							}
// 						}
// 					}
// 				}
// 			}
// 		}
// 	})

// 	c.OnError(func(r *colly.Response, err error) {
// 		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
// 	})
// 	if err := c.Visit(searchURL); err != nil {
// 		return nil, err
// 	}
// 	return imageURLs, nil
// }


// func getImageURLs2(query string) ([]string, error) {
// 	var imageURLs []string
// 	// var modifyImages []string
// 	var cookies []*http.Cookie

// 	c0 := colly.NewCollector()
// 	c0.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99 Safari/537.36"
// 	c0.OnResponse(func(r *colly.Response) {
// 		// Access the cookies
// 		cookies = c0.Cookies(r.Request.URL.String())
// 		for _, cookie := range cookies {
// 			// log.Println("Cookie:", cookie.Name, "Value:", cookie.Value)
// 			if cookie.Name == "SRCHHPGUSR" {
// 				cookie.Value = cookie.Value + "&ADLT=OFF"
// 			}
// 		}
// 	})

// 	if err := c0.Visit("https://www.bing.com/"); err != nil {
// 		return nil, err
// 	}

// 	// Bing画像検索のURL
// 	searchURL := fmt.Sprintf("https://www.bing.com/images/search?q=%s&qft=+filterui:aspect-tall", url.QueryEscape(query))

// 	fmt.Println("searchURL:", searchURL)
// 	// Collyのインスタンスを作成
// 	c := colly.NewCollector()
// 	// ユーザーエージェントを設定
// 	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99 Safari/537.36"

// 	c.OnRequest(func(r *colly.Request) {
// 		c.SetCookies("http://www.bing.com", cookies)
// 	})
// 	c.OnHTML("div.imgpt > a.iusc", func(e *colly.HTMLElement) {
// 		// ページから href を取得
// 		href := e.Attr("href")
// 		fmt.Println("Found href:", href)
// 		fmt.Println("Image src :", getCollectImageUrl2(href))
// 		imageURL := getCollectImageUrl2(href)
// 		if imageURL != "" {
// 			imageURLs = append(imageURLs, imageURL)
// 		}
// 	})
// 	c.OnError(func(r *colly.Response, err error) {
// 		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
// 	})
// 	if err := c.Visit(searchURL); err != nil {
// 		return nil, err
// 	}
// 	return imageURLs, nil
// }

// func getCollectImageUrl2(str string) string {
// 	re := regexp.MustCompile(`/images/search\?.*?mediaurl=(.*\.jpg)|/images/search\?.*?mediaurl=(.*\.png)|/images/search\?.*?mediaurl=([^&]+)`)
//     match := re.FindStringSubmatch(str)
// 	if len(match) > 1 {
// 		decodedURL, err := url.QueryUnescape(match[1])
// 		if err != nil {
// 			return ""
// 		}

// 		fmt.Println("decodedURL:", decodedURL)
// 		return decodedURL
// 	} else {
// 		return ""
// 	}
// }



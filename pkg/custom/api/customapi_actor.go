package api

import (
	// "encoding/json"
	// "fmt"
	"net/http"
	// "net/url"
	// "regexp"
	// "strings"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/xbapps/xbvr/pkg/models"

	// "github.com/gocolly/colly/v2"
)


type ResponseGetActors struct {
	Results int            `json:"results"`
	Scenes  []models.Actor `json:"actors"`
}

type ActorResource struct{}

func (i ActorResource) WebService() *restful.WebService {
	tags := []string{"CustomActor"}

	ws := new(restful.WebService)

	ws.Path("/api_custom/actor").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/addimages").To(i.addActorImages).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(models.Actor{}))

	ws.Route(ws.POST("/setfaceimage").To(i.setActorFaceImage).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(models.Actor{}))

	return ws
}

type RequestSetActorImage struct {
	ActorID uint   `json:"actor_id"`
	Url     string `json:"url"`
}

type RequestSetActorImages struct {
	ActorID uint   `json:"actor_id"`
	Urls     []string `json:"urls"`
}
// type RequestSearchActorImage struct {
// 	ActorID uint   `json:"actor_id"`
// 	Url     string `json:"url"`
// 	Keyword string `json:"keyword"`
// 	Site string    `json:"site"`
// }

func (i ActorResource) addActorImages(req *restful.Request, resp *restful.Response) {
	var r RequestSetActorImages
	err := req.ReadEntity(&r)
	if err != nil {
		log.Error(err)
		return
	}

	if r.ActorID == 0 || len(r.Urls) == 0 {
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

	for _, url := range r.Urls {
		actor.AddToImageArray(url)
	}

	err = actor.Save()
	if err != nil {
		log.Error(err)
		return
	}

	resp.WriteHeaderAndEntity(http.StatusOK, actor)
}

func (i ActorResource) setActorFaceImage(req *restful.Request, resp *restful.Response) {
	var r RequestSetActorImage
	err := req.ReadEntity(&r)
	if err != nil {
		log.Error(err)
		return
	}

	if r.ActorID == 0 || r.Url == "" {
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
	actor.FaceImageUrl = r.Url
	actor.AddToImageArray(r.Url)
	actor.Save()

	aa := models.ActionActor{ActorID: actor.ID, ActionType: "setimage", Source: "edit_actor", ChangedColumn: "face_image_url", NewValue: actor.FaceImageUrl}
	aa.Save()
	resp.WriteHeaderAndEntity(http.StatusOK, actor)
}
// func (i ActorResource) searchActorImage(req *restful.Request, resp *restful.Response) {
// 	var r RequestSearchActorImage
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
	
// 	if r.Site == "g" {
// 		imageURLs, err = getImageURLsFromGoogleImage("(\"" + actor.Name + "\") " + r.Keyword)
// 	} else {
// 		imageURLs, err = getImageURLs("(\"" + actor.Name + "\") " + r.Keyword)
// 	}
	
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}
	
// 	if len(imageURLs) == 0 || r.ActorID == 0 {
// 		return
// 	}

// 	actor.ImageUrl = r.Url
// 	for _, imageURL := range imageURLs {
// 		actor.AddToImageArray(imageURL)
// 	}
// 	actor.Save()

// 	// aa := models.ActionActor{ActorID: actor.ID, ActionType: "setimage", Source: "edit_actor", ChangedColumn: "image_url", NewValue: actor.ImageUrl}
// 	aa := models.ActionActor{ActorID: actor.ID, ActionType: "setimage", Source: "edit_actor", ChangedColumn: "image_arr", NewValue: actor.ImageArr}
// 	aa.Save()
// 	resp.WriteHeaderAndEntity(http.StatusOK, actor)
// }

// // URLから指定されたクエリパラメータを削除する関数
// func removeQueryParams(u string, paramsToRemove ...string) (string, error) {
// 	parsedURL, err := url.Parse(u)
// 	if err != nil {
// 		return "", err
// 	}

// 	query := parsedURL.Query()
// 	for _, param := range paramsToRemove {
// 		query.Del(param)
// 	}

// 	// クエリパラメータを削除したURLを再構築
// 	parsedURL.RawQuery = query.Encode()
// 	return parsedURL.String(), nil
// }

// // 指定された文字列とURLの先頭部分が一致するかどうかを確認する関数
// func isURLStartingWith(s, prefix string) bool {
// 	// URLをパース
// 	u, err := url.Parse(s)
// 	if err != nil {
// 		return false // URLの解析に失敗した場合は一致しないとみなす
// 	}

// 	// スキーム部分を取得し、指定された文字列と一致するかどうかを確認
// 	scheme := strings.ToLower(u.Scheme)
// 	return strings.HasPrefix(scheme, strings.ToLower(prefix))
// }

// func getImageURLsFromGoogleImage(query string) ([]string, error) {
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


// func getImageURLs(query string) ([]string, error) {
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
// 		fmt.Println("Image src :", getCollectImageUrl(href))
// 		imageURL := getCollectImageUrl(href)
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

// func getCollectImageUrl(str string) string {
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



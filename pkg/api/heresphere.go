package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/dustin/go-humanize"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/jinzhu/gorm"
	"github.com/markphelps/optional"
	"github.com/tidwall/gjson"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"

	"github.com/xbapps/xbvr/pkg/config"
	customapi "github.com/xbapps/xbvr/pkg/custom/api"
	"github.com/xbapps/xbvr/pkg/models"
	"github.com/xbapps/xbvr/pkg/scrape"
	"github.com/xbapps/xbvr/pkg/tasks"
)

type HeresphereLibrary struct {
	Access  int                    `json:"access"`
	Library []HeresphereListScenes `json:"library"`
}

type HeresphereListScenes struct {
	Name string   `json:"name"`
	List []string `json:"list"`
}

type HeresphereVideo struct {
	Access               int                            `json:"access"`
	Title                string                         `json:"title"`
	Description          string                         `json:"description"`
	ThumbnailImage       string                         `json:"thumbnailImage"`
	ThumbnailVideo       string                         `json:"thumbnailVideo,omitempty"`
	DateReleased         string                         `json:"dateReleased"`
	DateAdded            string                         `json:"dateAdded"`
	DurationMilliseconds uint                           `json:"duration"`
	Rating               float64                        `json:"rating,omitempty"`
	IsFavorite           bool                           `json:"isFavorite"`
	Projection           string                         `json:"projection"`
	Stereo               string                         `json:"stereo"`
	FOV                  float64                        `json:"fov"`
	Lens                 string                         `json:"lens"`
	HspUrl               string                         `json:"hsp,omitempty"`
	Scripts              []HeresphereScript             `json:"scripts,omitempty"`
	Subtitles            []HeresphereSubtitles          `json:"subtitles,omitempty"`
	Tags                 []HeresphereTag                `json:"tags,omitempty"`
	Media                []HeresphereMedia              `json:"media"`
	AlphaPackedSettings  *HereSphereAlphaPackedSettings `json:"alphaPackedSettings,omitempty"`
	WriteFavorite        bool                           `json:"writeFavorite"`
	WriteRating          bool                           `json:"writeRating"`
	WriteTags            bool                           `json:"writeTags"`
	WriteHSP             bool                           `json:"writeHSP"`
}

type HeresphereScript struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type HeresphereSubtitles struct {
	Name     string `json:"name"`
	Language string `json:"language"`
	URL      string `json:"url"`
}

type HeresphereTag struct {
	Name              string  `json:"name"`
	StartMilliseconds float64 `json:"start,omitempty"`
	EndMilliseconds   float64 `json:"end,omitempty"`
	Track             *int    `json:"track,omitempty"`
	Rating            float64 `json:"rating,omitempty"`
}

type HeresphereMedia struct {
	Name    string             `json:"name"`
	Sources []HeresphereSource `json:"sources"`
}

type StringOrInt string
type HeresphereSource struct {
	Resolution StringOrInt `json:"resolution"`
	Height     int         `json:"height"`
	Width      int         `json:"width"`
	Size       int64       `json:"size"`
	URL        string      `json:"url"`
}

type HereSphereAlphaPackedSettings struct {
	DefaultSettings bool `json:"defaultSettings"` // not an actual setting, the struct does not get marshalled if it is empty
}

type HereSphereAuthRequest struct {
	Username         string           `json:"username"`
	Password         string           `json:"password"`
	Rating           *float64         `json:"rating"`
	IsFavorite       *bool            `json:"isFavorite"`
	Hsp              *string          `json:"hsp"`
	Tags             *[]HeresphereTag `json:"tags"`
	DeleteFiles      *bool            `json:"deleteFile"`
	NeedsMediaSource optional.Bool    `json:"needsMediaSource"`
}

var RequestBody []byte

func HeresphereAuthFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	RequestBody, _ = io.ReadAll(req.Request.Body)
	if isDeoAuthEnabled() {
		authState := 0
		var requestData HereSphereAuthRequest

		if err := json.Unmarshal(RequestBody, &requestData); err == nil {
			if requestData.Username != "" && requestData.Password != "" {
				cmpErr := bcrypt.CompareHashAndPassword([]byte(config.Config.Interfaces.DeoVR.Password), []byte(requestData.Password))
				if requestData.Username == config.Config.Interfaces.DeoVR.Username && cmpErr == nil {
					authState = 1
				} else {
					authState = -1
				}
			}
		}

		if authState != 1 {
			msg := "Login Required"
			if authState == -1 {
				msg = "Login Failed"
			}
			unauthLib := HeresphereLibrary{
				Access: authState,
				Library: []HeresphereListScenes{
					{
						Name: msg,
						List: nil,
					},
				},
			}
			resp.WriteHeaderAndEntity(http.StatusOK, unauthLib)
			return
		}
	}
	chain.ProcessFilter(req, resp)
}

func HeresphereResponseFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	if req.Request.URL.Path != "/heresphere/menu" {
		resp.AddHeader("HereSphere-JSON-Version", "1")
	}
	chain.ProcessFilter(req, resp)
}

type HeresphereResource struct{}

func (i HeresphereResource) WebService() *restful.WebService {
	tags := []string{"HereSphere"}

	ws := new(restful.WebService)

	ws.Path("/heresphere").
		Filter(HeresphereResponseFilter).
		Consumes(restful.MIME_JSON, "application/x-www-form-urlencoded", "*/*").
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/menu").To(i.getHeresphereMenu).
		Produces("text/html"))
	ws.Route(ws.GET("/mode/scenes").To(i.getHeresphereModeScenes).
		Produces("text/html"))

	ws.Route(ws.HEAD("/").To(i.getHeresphereLibrary))

	ws.Route(ws.GET("/").Filter(HeresphereAuthFilter).To(i.getHeresphereLibrary).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(DeoLibrary{}))
	ws.Route(ws.POST("/").Filter(HeresphereAuthFilter).To(i.getHeresphereLibrary).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(DeoLibrary{}))

	ws.Route(ws.HEAD("/collection").To(i.getHeresphereCollection))
	ws.Route(ws.GET("/collection").Filter(HeresphereAuthFilter).To(i.getHeresphereCollection).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Produces(restful.MIME_JSON, "text/html").
		Writes(HeresphereLibrary{}))
	ws.Route(ws.POST("/collection").Filter(HeresphereAuthFilter).To(i.getHeresphereCollection).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(HeresphereLibrary{}))

	ws.Route(ws.HEAD("/collection/pickup").To(i.getHeresphereCollectionPickup))
	ws.Route(ws.GET("/collection/pickup").Filter(HeresphereAuthFilter).To(i.getHeresphereCollectionPickup).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Produces(restful.MIME_JSON, "text/html").
		Writes(HeresphereLibrary{}))
	ws.Route(ws.POST("/collection/pickup").Filter(HeresphereAuthFilter).To(i.getHeresphereCollectionPickup).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(HeresphereLibrary{}))

	ws.Route(ws.HEAD("/collection/ranking").To(i.getHeresphereCollectionRanking))
	ws.Route(ws.GET("/collection/ranking").Filter(HeresphereAuthFilter).To(i.getHeresphereCollectionRanking).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Produces(restful.MIME_JSON, "text/html").
		Writes(HeresphereLibrary{}))
	ws.Route(ws.POST("/collection/ranking").Filter(HeresphereAuthFilter).To(i.getHeresphereCollectionRanking).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(HeresphereLibrary{}))

	ws.Route(ws.HEAD("/collection/scene-ranking").To(i.getHeresphereCollectionSceneRanking))
	ws.Route(ws.GET("/collection/scene-ranking").Filter(HeresphereAuthFilter).To(i.getHeresphereCollectionSceneRanking).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Produces(restful.MIME_JSON, "text/html").
		Writes(HeresphereLibrary{}))
	ws.Route(ws.POST("/collection/scene-ranking").Filter(HeresphereAuthFilter).To(i.getHeresphereCollectionSceneRanking).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(HeresphereLibrary{}))

	ws.Route(ws.HEAD("/collection2").To(i.getHeresphereCollection2))
	ws.Route(ws.GET("/collection2").Filter(HeresphereAuthFilter).To(i.getHeresphereCollection2).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Produces(restful.MIME_JSON, "text/html").
		Writes(HeresphereLibrary{}))
	ws.Route(ws.POST("/collection2").Filter(HeresphereAuthFilter).To(i.getHeresphereCollection2).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(HeresphereLibrary{}))

	ws.Route(ws.HEAD("/actor/{actor-id}").To(i.getHeresphereActor))
	ws.Route(ws.GET("/actor/{actor-id}").Filter(HeresphereAuthFilter).To(i.getHeresphereActor).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(HeresphereVideo{}))
	ws.Route(ws.POST("/actor/{actor-id}").Filter(HeresphereAuthFilter).To(i.getHeresphereActor).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(HeresphereVideo{}))

	ws.Route(ws.GET("/{scene-id}").Filter(HeresphereAuthFilter).To(i.getHeresphereScene).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(DeoScene{}))
	ws.Route(ws.POST("/{scene-id}").Filter(HeresphereAuthFilter).To(i.getHeresphereScene).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(DeoScene{}))

	ws.Route(ws.GET("/file/{file-id}").Filter(HeresphereAuthFilter).To(i.getHeresphereFile).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(DeoScene{}))
	ws.Route(ws.POST("file/{file-id}").Filter(HeresphereAuthFilter).To(i.getHeresphereFile).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(DeoScene{}))
	return ws
}

func (i HeresphereResource) getHeresphereFile(req *restful.Request, resp *restful.Response) {
	if !config.Config.Interfaces.DeoVR.Enabled {
		return
	}

	var requestData HereSphereAuthRequest
	if err := json.Unmarshal(RequestBody, &requestData); err != nil {
		log.Warnf("Error decoding heresphere api POST request: %v %s", err, req.Request.RequestURI)
	}

	dnt := ""
	if !config.Config.Interfaces.DeoVR.TrackWatchTime {
		dnt = "?dnt=true"
	}

	db, _ := models.GetDB()
	defer db.Close()

	fileId, err := strconv.Atoi(req.PathParameter("file-id"))
	if err != nil {
		return
	}

	var file models.File
	db.Where(&models.File{ID: uint(fileId)}).First(&file)

	resolution := strconv.Itoa(file.VideoHeight)
	height := file.VideoHeight
	width := file.VideoWidth
	if file.VideoProjection == "360_tb" {
		resolution = strconv.Itoa(file.VideoHeight / 2)
	}

	var media []HeresphereMedia
	media = append(media, HeresphereMedia{
		Name: fmt.Sprintf("File 1/1 %vp - %v", resolution, humanize.Bytes(uint64(file.Size))),
		Sources: []HeresphereSource{
			{
				Resolution: StringOrInt(resolution),
				Height:     height,
				Width:      width,
				Size:       file.Size,
				URL:        fmt.Sprintf("%v://%v/api/dms/file/%v%v", getProto(req), req.Request.Host, file.ID, dnt),
			},
		},
	})

	video := HeresphereVideo{
		Access:               1,
		Title:                file.Filename,
		Description:          file.Filename,
		ThumbnailImage:       getProto(req) + "://" + req.Request.Host + "/ui/images/blank.png",
		DateReleased:         file.CreatedTime.Format("2006-01-02"),
		DateAdded:            file.CreatedTime.Format("2006-01-02"),
		DurationMilliseconds: uint(file.VideoDuration * 1000),
		Media:                media,
	}
	if requestData.DeleteFiles != nil && config.Config.Interfaces.Heresphere.AllowFileDeletes {
		log.Infof("Got request by HereSphere to delete file %v", file.Filename)
		removeFileByFileId(file.ID)
	}

	resp.WriteHeaderAndEntity(http.StatusOK, video)
}

func (i HeresphereResource) getHeresphereScene(req *restful.Request, resp *restful.Response) {
	if !config.Config.Interfaces.DeoVR.Enabled {
		return
	}

	var requestData HereSphereAuthRequest

	if err := json.Unmarshal(RequestBody, &requestData); err != nil {
		log.Warnf("Error decoding heresphere api POST request: %v %s", err, req.Request.RequestURI)
	}

	sceneID := req.PathParameter("scene-id")
	if sceneID == "" {
		return
	}

	dnt := ""
	if !config.Config.Interfaces.DeoVR.TrackWatchTime {
		dnt = "?dnt=true"
	}

	db, _ := models.GetDB()
	defer db.Close()

	var scene models.Scene
	err := db.Preload("Cast").
		Preload("Tags").
		Preload("Cuepoints", "track is not null").
		Preload("Files").
		Where("id = ?", sceneID).First(&scene).Error
	if err != nil {
		log.Error(err)
		return
	}
	if len(scene.Cuepoints) == 0 {
		db.Preload("Cast").
			Preload("Tags").
			Preload("Cuepoints", "track is null").
			Preload("Files").
			Where("id = ?", sceneID).First(&scene)
	}

	var videoFiles []models.File
	videoFiles, err = scene.GetVideoFilesSorted(config.Config.Interfaces.Players.VideoSortSeq)
	if err != nil {
		log.Error(err)
		return
	}

	if len(videoFiles) == 0 {
		ProcessHeresphereUpdates(&scene, requestData, models.File{})
	} else {
		ProcessHeresphereUpdates(&scene, requestData, videoFiles[0])
	}

	features := make(map[string]bool, 30)
	addFeatureTag := func(feature string) {
		if !features[feature] {
			features[feature] = true
		}
	}

	var media []HeresphereMedia

	videoLength := float64(scene.Duration)

	for i, file := range videoFiles {
		height := file.VideoHeight
		width := file.VideoWidth
		resolution := strconv.Itoa(file.VideoHeight)
		vertresolution := file.VideoWidth

		if file.VideoProjection == "360_tb" {
			resolution = strconv.Itoa(file.VideoHeight / 2)
			vertresolution = vertresolution * 2
		}

		resolutionClass := fmt.Sprintf("%0.fK", math.Round(float64(vertresolution)/1000))
		addFeatureTag("Resolution: " + resolutionClass)

		if file.VideoAvgFrameRateVal > 1.0 {
			addFeatureTag(fmt.Sprintf("Frame Rate: %.0ffps", file.VideoAvgFrameRateVal))
		}

		mediafile := HeresphereMedia{
			Name: fmt.Sprintf("File %v/%v %vp - %v", i+1, len(videoFiles), resolution, humanize.Bytes(uint64(file.Size))),
			Sources: []HeresphereSource{
				{
					Resolution: StringOrInt(resolution),
					Height:     height,
					Width:      width,
					Size:       file.Size,
					URL:        fmt.Sprintf("%v://%v/api/dms/file/%v%v", getProto(req), req.Request.Host, file.ID, dnt),
				},
			},
		}

		media = append(media, mediafile)
		videoLength = file.VideoDuration
	}

	if len(videoFiles) == 0 && config.Config.Web.SceneTrailerlist && requestData.NeedsMediaSource.OrElse(true) {
		switch scene.TrailerType {
		case "heresphere":
			heresphereScene := LoadHeresphereScene(scene.TrailerSource)
			media = append(media, heresphereScene.Media...)
		case "slr", "deovr":
			deovrScene := LoadDeovrScene(scene.TrailerSource)
			var hsp HeresphereMedia
			for _, encoding := range deovrScene.Encodings {
				if len(encoding.VideoSources) > 0 {
					hsp.Name = encoding.Name
					for _, source := range encoding.VideoSources {
						hspSource := HeresphereSource(HeresphereSource{Resolution: StringOrInt(strconv.Itoa(source.Resolution)), Height: source.Height, Width: source.Width, Size: source.Size, URL: source.URL})
						hsp.Sources = append(hsp.Sources, hspSource)
					}
					media = append(media, hsp)
				}
			}

		case "url":
			sources := LoadUrl(scene.TrailerSource)
			media = copyVideoSourceResponse(sources, media)
		case "urls":
			sources := models.VideoSourceResponse{}
			if err := json.Unmarshal([]byte(scene.TrailerSource), &sources); err == nil {
				media = copyVideoSourceResponse(sources, media)
			}
		case "scrape_html":
			sources := ScrapeHtml(scene.TrailerSource)
			media = copyVideoSourceResponse(sources, media)
		case "scrape_json":
			sources := ScrapeJson(scene.TrailerSource)
			media = copyVideoSourceResponse(sources, media)
		case "load_json":
			sources := LoadJson(scene.TrailerSource)
			media = copyVideoSourceResponse(sources, media)
		case "vrporn":
			sources := scrape.VRPornTrailer(scene.TrailerSource)
			media = copyVideoSourceResponse(sources, media)
		}
	}

	if len(videoFiles) > 1 {
		addFeatureTag("Multiple video files")
	}

	var tags []HeresphereTag

	cuepoints := scene.Cuepoints
	sort.Slice(cuepoints, func(i, j int) bool {
		return cuepoints[i].TimeStart < cuepoints[j].TimeStart
	})

	if len(scene.Cuepoints) > 0 {
		if scene.Cuepoints[0].Track == nil {
			// basic cuepoint data
			end := 0
			var trackAssignments []string
			for i := range cuepoints {
				start := int(cuepoints[i].TimeStart * 1000)
				if i+1 < len(cuepoints) {
					end = int(cuepoints[i+1].TimeStart*1000) - 1
				} else if int(videoLength*1000) > start {
					end = int(videoLength * 1000)
				} else {
					end = start + 1000
				}

				if !config.Config.Interfaces.Heresphere.MultitrackCuepoints && !config.Config.Interfaces.Heresphere.MultitrackCastCuepoints {
					track := 0
					tags = append(tags, HeresphereTag{
						Name:              scene.Cuepoints[i].Name,
						StartMilliseconds: float64(start),
						EndMilliseconds:   float64(end),
						Track:             &track,
					})
				} else {
					// determine cast ve non-cast cuepoints
					var cast []string
					var cuepoints []string
					for _, cuepointName := range strings.Split(scene.Cuepoints[i].Name, "-") {
						if isCast(cuepointName, scene) {
							cast = append(cast, cuepointName)
						} else {
							cuepoints = append(cuepoints, cuepointName)
						}
					}

					if config.Config.Interfaces.Heresphere.MultitrackCuepoints {
						for idx, cuepoint := range cuepoints {
							track, _ := findTrack(idx, len(cuepoints), cuepoint, &trackAssignments, scene)
							if track == 0 && !config.Config.Interfaces.Heresphere.MultitrackCastCuepoints && len(cast) > 0 {
								cuepoint = strings.Join(cast, "-") + "-" + cuepoint
							}
							tags = append(tags, HeresphereTag{
								Name:              cuepoint,
								StartMilliseconds: float64(start),
								EndMilliseconds:   float64(end),
								Track:             &track,
							})
						}
					} else {
						track := 0
						tags = append(tags, HeresphereTag{
							Name:              strings.Join(cuepoints, "-"),
							StartMilliseconds: float64(start),
							EndMilliseconds:   float64(end),
							Track:             &track,
						})
					}
					if len(cast) > 0 {
						if config.Config.Interfaces.Heresphere.MultitrackCastCuepoints {
							for _, actor := range cast {
								track, _ := findTrack(0, 0, actor, &trackAssignments, scene)
								tags = append(tags, HeresphereTag{
									Name:              actor,
									StartMilliseconds: float64(start),
									EndMilliseconds:   float64(end),
									Track:             &track,
								})
							}
						} else {
							if len(cuepoints) == 0 {
								track := 0
								tags = append(tags, HeresphereTag{
									Name:              strings.Join(cast, "-"),
									StartMilliseconds: float64(start),
									EndMilliseconds:   float64(end),
									Track:             &track,
								})
							}
						}
					}
				}
			}
		} else {
			// extra heresphere cuepoint attributes
			for _, cuepoint := range cuepoints {
				trackNo := int(*cuepoint.Track)
				tags = append(tags, HeresphereTag{
					Name:              cuepoint.Name,
					StartMilliseconds: cuepoint.TimeStart * 1000,
					EndMilliseconds:   cuepoint.TimeEnd * 1000,
					Track:             &trackNo,
					Rating:            cuepoint.Rating,
				})

			}
		}
		addFeatureTag("Has cuepoints")
	}

	if scene.IsHidden {
		addFeatureTag("Hidden")
	}

	tags = append(tags, HeresphereTag{
		Name: "Studio:" + scene.Site,
	})

	akaCnt := 0
	for i := range scene.Cast {
		if strings.HasPrefix(scene.Cast[i].Name, "aka:") {
			akaCnt++
			tags = append(tags, HeresphereTag{
				Name: strings.Replace(scene.Cast[i].Name, ",", "/", -1),
			})
		} else {
			tags = append(tags, HeresphereTag{
				Name: "Talent:" + scene.Cast[i].Name,
			})
		}
	}
	if (len(scene.Cast) - akaCnt) > 5 {
		addFeatureTag("Cast: 6+")
	} else if (len(scene.Cast) - akaCnt) > 0 {
		addFeatureTag(fmt.Sprintf("Cast: %d", (len(scene.Cast) - akaCnt)))
	}

	for i := range scene.Tags {
		if strings.HasPrefix(scene.Tags[i].Name, "tag group:") {
			akaCnt++
			tags = append(tags, HeresphereTag{
				Name: strings.Replace(strings.Replace(scene.Tags[i].Name, ",", "/", -1), "tag group:", "Tag Group:", 1),
			})
		} else {
			tags = append(tags, HeresphereTag{
				Name: "Category:" + scene.Tags[i].Name,
			})
		}
	}

	var heresphereScriptFiles []HeresphereScript
	var scriptFiles []models.File
	scriptFiles, err = scene.GetScriptFilesSorted(config.Config.Interfaces.Players.ScriptSortSeq)
	if err != nil {
		log.Error(err)
		return
	}

	for _, file := range scriptFiles {
		addFeatureTag("Is scripted")
		heresphereScriptFiles = append(heresphereScriptFiles, HeresphereScript{
			Name: file.Filename,
			URL:  fmt.Sprintf("%v://%v/api/dms/file/%v", getProto(req), req.Request.Host, file.ID),
		})

		if scene.HumanScript {
			addFeatureTag("Hand Crafted Funscript")
		}
		if scene.AiScript {
			addFeatureTag("AI Generated Funscript")
		}
	}

	var heresphereSubtitlesFiles []HeresphereSubtitles
	var subtitlesFiles []models.File
	subtitlesFiles, err = scene.GetSubtitlesFilesSorted(config.Config.Interfaces.Players.SubtitleSortSeq)
	if err != nil {
		log.Error(err)
		return
	}

	getLanguage := func(path string) string {
		path = path[:len(path)-4]
		index := strings.LastIndex(path, ".")
		if index > -1 {
			languageCode := path[index+1:]
			languageTag, err := language.Parse(languageCode)
			if err != nil {
				log.Error(err)
				languageTag = language.MustParse("en")
			}
			en := display.English.Languages()
			return en.Name(languageTag)
		} else {
			languageTag := language.MustParse("en")
			en := display.English.Languages()
			return en.Name(languageTag)
		}
	}

	for _, file := range subtitlesFiles {
		addFeatureTag("Has subtitles")
		heresphereSubtitlesFiles = append(heresphereSubtitlesFiles, HeresphereSubtitles{
			Name:     file.Filename,
			Language: getLanguage(file.Filename),
			URL:      fmt.Sprintf("%v://%v/api/dms/file/%v", getProto(req), req.Request.Host, file.ID),
		})
	}

	hspUrl := ""
	var hspFiles []models.File
	hspFiles, err = scene.GetHSPFiles()
	if err != nil {
		log.Error(err)
		return
	}

	if len(hspFiles) > 0 {
		addFeatureTag("Has HSP file")
		hspUrl = fmt.Sprintf("%v://%v/api/dms/file/%v", getProto(req), req.Request.Host, hspFiles[0].ID)
	}

	var projection string = "equirectangular"
	var stereo string = "sbs"
	fov := 180.0
	lens := "Linear"

	if len(videoFiles) == 0 {
		videoFiles = append(videoFiles, models.File{})
	}
	switch videoFiles[0].VideoProjection {
	case "flat":
		addFeatureTag("Flat video")
		projection = "perspective"
		stereo = "mono"

	case "180_mono":
		addFeatureTag("FOV: 180°")
		projection = "equirectangular"
		stereo = "mono"

	case "360_mono":
		addFeatureTag("FOV: 360°")
		projection = "equirectangular360"
		stereo = "mono"

	case "180_sbs":
		addFeatureTag("FOV: 180°")
		projection = "equirectangular"

	case "360_tb":
		addFeatureTag("FOV: 360°")
		projection = "equirectangular360"
		stereo = "tb"

	case "mkx200":
		addFeatureTag("FOV: 200°")
		projection = "fisheye"
		fov = 200.0
		lens = "MKX200"

	case "mkx220":
		addFeatureTag("FOV: 220°")
		projection = "fisheye"
		fov = 220.0
		lens = "MKX220"

	case "vrca220":
		addFeatureTag("FOV: 220°")
		projection = "fisheye"
		fov = 220.0
		lens = "VRCA220"

	case "rf52":
		addFeatureTag("FOV: 190°")
		projection = "fisheye"
		fov = 190.0

	case "fisheye190":
		addFeatureTag("FOV: 190°")
		projection = "fisheye"
		fov = 190.0

	case "fisheye":
		addFeatureTag("FOV: 180°")
		projection = "fisheye"
	}

	title := scene.Title
	thumbnailURL := getProto(req) + "://" + req.Request.Host + "/img/700x/" + strings.Replace(scene.CoverURL, "://", ":/", -1)

	if scene.IsScripted {
		title = scene.GetFunscriptTitle()
		if config.Config.Interfaces.DeoVR.RenderHeatmaps {
			thumbnailURL = getProto(req) + "://" + req.Request.Host + "/imghm/" + fmt.Sprint(scene.ID) + "/" + strings.Replace(scene.CoverURL, "://", ":/", -1)
		}
	}

	if scene.Watchlist {
		addFeatureTag("Watchlist")
	}
	if scene.Trailerlist {
		addFeatureTag("Trailer List")
	}

	if scene.ReleaseDate.Year() > 1900 {
		addFeatureTag("Month: " + scene.ReleaseDate.Format("2006-01"))
		addFeatureTag("Year: " + scene.ReleaseDate.Format("2006"))
	}

	var alphaPackedSettings *HereSphereAlphaPackedSettings = nil
	if videoFiles[0].HasAlpha {
		alphaPackedSettings = &HereSphereAlphaPackedSettings{DefaultSettings: true}
		addFeatureTag("Is Alpha Passthrough")
	} else if gjson.Valid(scene.ChromaKey) {
		result := gjson.Get(scene.ChromaKey, "hasAlpha")
		if result.Exists() && result.Bool() {
			alphaPackedSettings = &HereSphereAlphaPackedSettings{DefaultSettings: true}
			addFeatureTag("Is Alpha Passthrough")
		}
	}

	for f := range features {
		tags = append(tags, HeresphereTag{
			Name: "Feature:" + f,
		})
	}

	video := HeresphereVideo{
		Access:               1,
		Title:                title,
		Description:          scene.Synopsis,
		ThumbnailImage:       thumbnailURL,
		DateReleased:         scene.ReleaseDate.Format("2006-01-02"),
		DateAdded:            scene.AddedDate.Format("2006-01-02"),
		DurationMilliseconds: uint(videoLength * 1000),
		Rating:               scene.StarRating,
		IsFavorite:           scene.Favourite,
		Projection:           projection,
		Stereo:               stereo,
		FOV:                  fov,
		Lens:                 lens,
		HspUrl:               hspUrl,
		Scripts:              heresphereScriptFiles,
		Subtitles:            heresphereSubtitlesFiles,
		Tags:                 tags,
		Media:                media,
		AlphaPackedSettings:  alphaPackedSettings,
		WriteFavorite:        config.Config.Interfaces.Heresphere.AllowFavoriteUpdates,
		WriteRating:          config.Config.Interfaces.Heresphere.AllowRatingUpdates,
		WriteTags:            config.Config.Interfaces.Heresphere.AllowTagUpdates || config.Config.Interfaces.Heresphere.AllowCuepointUpdates || config.Config.Interfaces.Heresphere.AllowWatchlistUpdates || config.Config.Web.SceneTrailerlist,
		WriteHSP:             config.Config.Interfaces.Heresphere.AllowHspData,
	}

	if scene.HasVideoPreview {
		video.ThumbnailVideo = fmt.Sprintf("%v://%v/api/dms/preview/%v", getProto(req), req.Request.Host, scene.SceneID)
	}

	resp.WriteHeaderAndEntity(http.StatusOK, video)
}

func copyVideoSourceResponse(sources models.VideoSourceResponse, media []HeresphereMedia) []HeresphereMedia {
	if len(sources.VideoSources) > 0 {
		for _, source := range sources.VideoSources {
			var hsp HeresphereMedia
			hsp.Name = source.Quality
			hspSource := HeresphereSource{
				URL: source.URL,
			}
			hsp.Sources = append(hsp.Sources, hspSource)
			media = append(media, hsp)
		}
	}
	return media
}

var lockHeresphereUpdates sync.Mutex

func ProcessHeresphereUpdates(scene *models.Scene, requestData HereSphereAuthRequest, videoFile models.File) {
	db, _ := models.GetDB()
	defer db.Close()

	if requestData.IsFavorite != nil && *requestData.IsFavorite != scene.Favourite && config.Config.Interfaces.Heresphere.AllowFavoriteUpdates {
		scene.Favourite = *requestData.IsFavorite
		scene.Save()
	}
	if requestData.Rating != nil && *requestData.Rating != scene.StarRating && config.Config.Interfaces.Heresphere.AllowRatingUpdates {
		scene.StarRating = *requestData.Rating
		scene.Save()
	}

	if requestData.Tags != nil && (config.Config.Interfaces.Heresphere.AllowTagUpdates || config.Config.Interfaces.Heresphere.AllowCuepointUpdates || config.Config.Interfaces.Heresphere.AllowWatchlistUpdates || config.Config.Web.SceneTrailerlist) {
		// need lock, heresphere can send a second post too soon
		lockHeresphereUpdates.Lock()
		defer lockHeresphereUpdates.Unlock()
	}
	if requestData.Tags != nil && config.Config.Interfaces.Heresphere.AllowTagUpdates {
		var newTags []string

		// need to reread the tags, to handle muti threading issues and the scene record may have changed
		// just preload the tags, preload all associations and the scene, does not reread the tags?, so just get them and update the scene
		var tmp models.Scene
		db.Preload("Tags").Where("id = ?", scene.ID).First(&tmp)
		scene.Tags = tmp.Tags

		for _, tag := range *requestData.Tags {
			if strings.HasPrefix(strings.ToLower(tag.Name), "category:") {
				newTags = append(newTags, tag.Name[9:])
			}
		}
		ProcessTagChanges(scene, &newTags, db)
		scene.Save()
	}

	if requestData.Tags != nil && (config.Config.Interfaces.Heresphere.AllowWatchlistUpdates || config.Config.Web.SceneTrailerlist) {
		// need to reread the tags, to handle muti threading issues and the scene record may have changed
		// just preload the tags, preload all associations and the scene, does not reread the tags?, so just get them and update the scene
		var tmp models.Scene
		db.Preload("Tags").Where("id = ?", scene.ID).First(&tmp)
		scene.Tags = tmp.Tags

		watchlist := false
		trailerlist := false
		for _, tag := range *requestData.Tags {
			if strings.HasPrefix(strings.ToLower(tag.Name), "feature:watchlist") {
				watchlist = true
			}
			if strings.HasPrefix(strings.ToLower(tag.Name), "feature:trailer list") {
				trailerlist = true
			}
		}
		if scene.Watchlist != watchlist && config.Config.Interfaces.Heresphere.AllowWatchlistUpdates {
			scene.Watchlist = watchlist
			scene.Save()
		}
		if scene.Trailerlist != trailerlist && config.Config.Web.SceneTrailerlist {
			scene.Trailerlist = trailerlist
			scene.Save()
		}
	}

	if requestData.Tags != nil && config.Config.Interfaces.Heresphere.AllowCuepointUpdates {
		// need to reread the cuepoints, to handle muti threading issues and the scene record may have changed
		// just preload the cuepoint, preload all associations and the scene, does not reread the cuepoint?, so just get them and update the scene
		var existingScene models.Scene
		db.Preload("Cuepoints", "track is not null").Where("id = ?", scene.ID).First(&existingScene)

		var replacementCuepoints []models.SceneCuepoint
		endpos := findEndPos(requestData)
		firstTrack := findTheMainTrack(requestData)
		// build new list of cuepoints
		newTrack := -1
		lastTrack := -1
		for _, tag := range *requestData.Tags {
			if !strings.Contains(tag.Name, ":") || (strings.HasPrefix(tag.Name, "Talent:") && tag.StartMilliseconds != 0 && tag.EndMilliseconds != endpos) {
				if lastTrack != int(*tag.Track) {
					// adjust track numbers, starting at track 0
					lastTrack = int(*tag.Track)
					newTrack += 1
				}
				*tag.Track = newTrack
				if strings.HasPrefix(tag.Name, "Talent:") {
					// allows users to create a Cuepoint with the Talent: tags, but remove the prefix
					tag.Name = strings.Replace(tag.Name, "Talent:", "", 1)
				}
				if *tag.Track <= firstTrack+1000 {
					var trackNo uint
					if tag.Track != nil {
						trackNo = uint(*tag.Track)
					}
					replacementCuepoints = append(replacementCuepoints, models.SceneCuepoint{SceneID: scene.ID, TimeStart: float64(tag.StartMilliseconds) / 1000, Name: tag.Name, Track: &trackNo, TimeEnd: float64(tag.EndMilliseconds) / 1000, Rating: tag.Rating})
				} else {
					// allow for multi track, merge into the main cuepoint name
					if tag.StartMilliseconds > 0 || tag.EndMilliseconds < endpos {
						for idx, newtag := range replacementCuepoints {
							// allow 5 seconds lewway to align manually entered tags
							if math.Abs((newtag.TimeStart)-tag.StartMilliseconds/1000) < 5 {
								replacementCuepoints[idx].Name = tag.Name + "-" + replacementCuepoints[idx].Name
							}
						}
					}
				}
			}
		}

		// delete existing HSP cuepoints (track not null)
		var emptyList []models.SceneCuepoint
		scene.Cuepoints = emptyList
		db.Where("scene_id = ? and track is not null", scene.ID).Delete(&models.SceneCuepoint{})

		// readd new set
		for _, newCuepoint := range replacementCuepoints {
			db.Model(&scene).Association("Cuepoints").Append(&newCuepoint)
		}

		// delete non-HSP cuepoints (track is null) if the option is set
		if len(scene.Cuepoints) > 0 && !config.Config.Interfaces.Heresphere.RetainNonHSPCuepoints {
			db.Where("scene_id = ? and track is null", scene.ID).Delete(&models.SceneCuepoint{})
		}
	}

	if requestData.DeleteFiles != nil && config.Config.Interfaces.Heresphere.AllowFileDeletes {
		log.Infof("Got request by HereSphere to delete files for scene %v", scene.ID)
		for _, sceneFile := range scene.Files {
			removeFileByFileId(sceneFile.ID)
		}
	}

	if requestData.Hsp != nil && config.Config.Interfaces.Heresphere.AllowHspData {
		hspContent, err := base64.StdEncoding.DecodeString(*requestData.Hsp)
		if err != nil {
			log.Errorf("Error decoding heresphere hsp data %v", err)
		}

		fName := filepath.Join(scene.Files[0].Path, strings.TrimSuffix(scene.Files[0].Filename, filepath.Ext(videoFile.Filename))+".hsp")
		os.WriteFile(fName, hspContent, 0644)

		tasks.ScanLocalHspFile(fName, videoFile.VolumeID, scene.ID)
	}
}

func findTheMainTrack(requestData HereSphereAuthRequest) int {
	// 99% of the time we want Track 0, but the user may have deleted and added whole track

	// find the max duration
	endpos := findEndPos(requestData)
	for _, tag := range *requestData.Tags {
		if endpos < tag.EndMilliseconds {
			endpos = tag.EndMilliseconds
		}
	}

	// find the best track
	likelyTrack := 9999
	alternateTrack := 9999

	for _, tag := range *requestData.Tags {
		if (tag.StartMilliseconds > 0 || tag.EndMilliseconds < endpos) && !strings.Contains(tag.Name, ":") {
			return *tag.Track
		}

		if (tag.StartMilliseconds > 0 || tag.EndMilliseconds < endpos) && likelyTrack > *tag.Track {
			likelyTrack = *tag.Track
		}
		if !strings.Contains(tag.Name, ":") && alternateTrack > *tag.Track {
			alternateTrack = *tag.Track
		}
	}

	if likelyTrack < 9999 {
		return likelyTrack
	}

	if alternateTrack < 9999 {
		return likelyTrack
	}

	return -1
}

func findEndPos(requestData HereSphereAuthRequest) float64 {
	// find the max duration
	endpos := float64(0)
	for _, tag := range *requestData.Tags {
		if endpos < tag.EndMilliseconds {
			endpos = tag.EndMilliseconds
		}
	}
	return endpos
}

func (i HeresphereResource) getHeresphereLibrary(req *restful.Request, resp *restful.Response) {
	log.Infof("getHeresphereLibrary called, enabled %v", config.Config.Interfaces.DeoVR.Enabled)
	if !config.Config.Interfaces.DeoVR.Enabled {
		return
	}

	switch getHeresphereMode(req) {
	case "collection":
		i.writeHeresphereActorLibrary(req, resp, models.RequestActorList{
			Sort:     optional.NewString("furigana_asc"),
			MinAvail: optional.NewInt(1),
			Limit:    optional.NewInt(10000),
		}, "", "collection", false)
		return
	case "pickup":
		i.writeHeresphereActorLibrary(req, resp, models.RequestActorList{
			Sort:     optional.NewString("random"),
			MinAvail: optional.NewInt(1),
			Limit:    optional.NewInt(50),
		}, "", "pickup", false)
		return
	case "ranking":
		i.writeHeresphereActorLibrary(req, resp, models.RequestActorList{
			Sort:     optional.NewString("rating_desc"),
			MinAvail: optional.NewInt(1),
			Limit:    optional.NewInt(50),
		}, "総合", "ranking", false)
		return
	case "scene-ranking":
		i.writeHeresphereActorLibrary(req, resp, models.RequestActorList{
			Sort:     optional.NewString("scene_rating_desc"),
			MinAvail: optional.NewInt(1),
			Limit:    optional.NewInt(50),
		}, "シーン総合", "scene-ranking", false)
		return
	case "collection2":
		i.writeHeresphereCollection2Library(req, resp, false)
		return
	}

	db, _ := models.GetDB()
	defer db.Close()

	var sceneLists []HeresphereListScenes

	var savedPlaylists []models.Playlist
	db.Where("is_deo_enabled = ?", true).Order("ordering asc").Find(&savedPlaylists)

	for i := range savedPlaylists {
		var r models.RequestSceneList

		if err := json.Unmarshal([]byte(savedPlaylists[i].SearchParams), &r); err == nil {
			r.IsAccessible = optional.NewBool(true)
			r.IsAvailable = optional.NewBool(true)

			list := models.QuerySceneIDs(r)

			for i := range list {
				list[i] = fmt.Sprintf("%v://%v/heresphere/%v", getProto(req), req.Request.Host, list[i])
			}

			sceneLists = append(sceneLists, HeresphereListScenes{
				Name: savedPlaylists[i].Name,
				List: list,
			})
		}
	}

	// Add unmatched files at the end
	var unmatched []models.File
	db.Model(&unmatched).
		Preload("Volume").
		Where("files.scene_id = 0").
		Where("files.type = 'video'").
		Order("created_time desc").
		Find(&unmatched)

	if len(unmatched) > 0 {
		list := make([]string, len(unmatched))
		for i := range unmatched {
			url := fmt.Sprintf("%v://%v/heresphere/file/%v", getProto(req), req.Request.Host, unmatched[i].ID)
			list[i] = url
		}

		sceneLists = append(sceneLists, HeresphereListScenes{
			Name: "Unmatched",
			List: list,
		})

	}

	if config.Config.Web.SceneTrailerlist {
		var trailerlist []models.Scene
		db.Preload("Tags").Where("is_available = false and trailer_source != '' and (trailerlist )").Order("updated_at desc").Find(&trailerlist)
		if len(trailerlist) > 0 {
			var list []string
			for i := range trailerlist {
				downloadTag := false
				for _, tag := range trailerlist[i].Tags {
					if tag.Name == "download" {
						downloadTag = true
					}
				}
				if !downloadTag {
					url := fmt.Sprintf("%v://%v/heresphere/%v", getProto(req), req.Request.Host, trailerlist[i].ID)
					list = append(list, url)
				}
			}

			sceneLists = append(sceneLists, HeresphereListScenes{
				Name: "Trailers",
				List: list,
			})

		}
	}

	resp.WriteHeaderAndEntity(http.StatusOK, HeresphereLibrary{
		Access:  1,
		Library: sceneLists,
	})
}

func (i HeresphereResource) getHeresphereMenu(req *restful.Request, resp *restful.Response) {
	proto := getProto(req)
	host := req.Request.Host
	base := fmt.Sprintf("%s://%s/heresphere", proto, host)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ja">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>XBVR - HereSphere Menu</title>
  <style>
    * { box-sizing: border-box; }
    body {
      margin: 0;
      padding: 28px 20px 40px;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Hiragino Kaku Gothic ProN", "Yu Gothic", sans-serif;
      background: radial-gradient(circle at 20%% 0%%, #162433 0%%, #0d141c 70%%);
      color: #f1f5f9;
      min-height: 100vh;
    }
    .wrap {
      max-width: 720px;
      margin: 0 auto;
    }
    h1 {
      margin: 0 0 8px;
      font-size: 26px;
      font-weight: 800;
      color: #ffffff;
    }
    .guide {
      margin: 0 0 24px;
      font-size: 14px;
      line-height: 1.6;
      color: #94a3b8;
    }
    .grid {
      display: flex;
      flex-direction: column;
      gap: 14px;
    }
    .card {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
      padding: 18px 22px;
      background: #16202a;
      border: 1px solid #2b3f54;
      border-radius: 12px;
      text-decoration: none;
      color: #f8fafc;
      transition: background 0.15s ease, border-color 0.15s ease, transform 0.12s ease;
    }
    .card:hover, .card:focus {
      background: #1e2d3d;
      border-color: #00a8e1;
      transform: translateY(-1px);
      outline: none;
    }
    .card-title {
      font-size: 19px;
      font-weight: 700;
      margin: 0 0 4px;
      color: #ffffff;
    }
    .card-desc {
      font-size: 13px;
      color: #94a3b8;
      margin: 0;
    }
    .card-url {
      font-size: 12px;
      color: #38bdf8;
      margin-top: 6px;
      font-family: monospace;
    }
    .arrow {
      font-size: 22px;
      font-weight: 700;
      color: #00a8e1;
      flex-shrink: 0;
    }
    .badge {
      display: inline-block;
      font-size: 11px;
      font-weight: 700;
      padding: 2px 8px;
      border-radius: 999px;
      margin-right: 8px;
      vertical-align: middle;
      background: rgba(0, 168, 225, 0.2);
      color: #38bdf8;
      border: 1px solid rgba(0, 168, 225, 0.45);
    }
  </style>
</head>
<body>
  <div class="wrap">
    <h1>XBVR HereSphere メニュー</h1>
    <p class="guide">
      利用したいリストをタップして開いた状態で、HereSphere の <strong>API 切替ボタン</strong> を押してください。<br />
      別のリストに切り替える場合はブラウザに戻って再度選択してください。
    </p>
    <div class="grid">
      <a class="card" href="%[1]s/mode/scenes">
        <div>
          <div class="card-title"><span class="badge">SCENE</span>従来のシーンリスト</div>
          <p class="card-desc">プレイリスト・未マッチ・トレーラー等の通常ライブラリを表示します</p>
          <div class="card-url">%[1]s</div>
        </div>
        <span class="arrow">&rsaquo;</span>
      </a>
      <a class="card" href="%[1]s/collection">
        <div>
          <div class="card-title"><span class="badge">ACTOR</span>コレクション：全女優リスト</div>
          <p class="card-desc">視聴可能シーンがある全女優をふりがな順で表示します</p>
          <div class="card-url">%[1]s/collection</div>
        </div>
        <span class="arrow">&rsaquo;</span>
      </a>
      <a class="card" href="%[1]s/collection/pickup">
        <div>
          <div class="card-title"><span class="badge">PICKUP</span>ピックアップ（おすすめ女優）</div>
          <p class="card-desc">視聴可能シーンがある女優からランダムで最大50名を表示します</p>
          <div class="card-url">%[1]s/collection/pickup</div>
        </div>
        <span class="arrow">&rsaquo;</span>
      </a>
      <a class="card" href="%[1]s/collection/ranking">
        <div>
          <div class="card-title"><span class="badge">RANK</span>ランキング：総合</div>
          <p class="card-desc">総合評価順の上位50名（総合1．女優名）を表示します</p>
          <div class="card-url">%[1]s/collection/ranking</div>
        </div>
        <span class="arrow">&rsaquo;</span>
      </a>
      <a class="card" href="%[1]s/collection/scene-ranking">
        <div>
          <div class="card-title"><span class="badge">RANK</span>ランキング：シーン総合</div>
          <p class="card-desc">シーン平均評価順の上位50名（シーン総合1．女優名）を表示します</p>
          <div class="card-url">%[1]s/collection/scene-ranking</div>
        </div>
        <span class="arrow">&rsaquo;</span>
      </a>
      <a class="card" href="%[1]s/collection2">
        <div>
          <div class="card-title"><span class="badge">ACTOR=SCENE</span>コレクション2（女優＝1シーン・全動画順次再生）</div>
          <p class="card-desc">1人の女優を1シーンとして表示し、出演する全動画ファイルを順次再生します</p>
          <div class="card-url">%[1]s/collection2</div>
        </div>
        <span class="arrow">&rsaquo;</span>
      </a>
    </div>
  </div>
</body>
</html>`, base)

	resp.Header().Set("Content-Type", "text/html; charset=utf-8")
	resp.WriteHeader(http.StatusOK)
	_, _ = resp.Write([]byte(html))
}

var (
	heresphereModeMu sync.RWMutex
	heresphereModes  = map[string]string{}
)

func getHeresphereClientIP(req *restful.Request) string {
	if xff := req.HeaderParameter("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(req.Request.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return req.Request.RemoteAddr
}

func setHeresphereMode(req *restful.Request, mode string) {
	ip := getHeresphereClientIP(req)
	heresphereModeMu.Lock()
	heresphereModes[ip] = mode
	heresphereModeMu.Unlock()
}

func getHeresphereMode(req *restful.Request) string {
	ip := getHeresphereClientIP(req)
	heresphereModeMu.RLock()
	mode := heresphereModes[ip]
	heresphereModeMu.RUnlock()
	return mode
}

func writeHeresphereReadyPage(req *restful.Request, resp *restful.Response, title string) {
	proto := getProto(req)
	host := req.Request.Host
	currentURL := fmt.Sprintf("%s://%s%s", proto, host, req.Request.URL.Path)
	menuURL := fmt.Sprintf("%s://%s/heresphere/menu", proto, host)
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ja">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>HereSphere API Ready - %s</title>
  <style>
    body { margin: 0; padding: 36px 24px; font-family: sans-serif; background: #0d141c; color: #f1f5f9; text-align: center; }
    .box { max-width: 560px; margin: 0 auto; background: #16202a; border: 1px solid #2b3f54; border-radius: 12px; padding: 28px 24px; }
    h1 { font-size: 22px; margin: 0 0 10px; color: #ffffff; }
    .mode { font-size: 16px; font-weight: 700; color: #38bdf8; margin: 0 0 14px; }
    p { font-size: 14px; color: #94a3b8; line-height: 1.6; margin: 0 0 20px; }
    .url { font-family: monospace; font-size: 13px; color: #64748b; word-break: break-all; margin-bottom: 20px; }
    .btn { display: inline-block; padding: 12px 22px; border-radius: 8px; background: #1e2d3d; border: 1px solid #00a8e1; color: #ffffff; text-decoration: none; font-weight: 700; font-size: 14px; }
  </style>
</head>
<body>
  <div class="box">
    <h1>HereSphere API 待機中</h1>
    <div class="mode">%s</div>
    <p>この状態で HereSphere の <strong>API 切替ボタン</strong> を押すと、選択したリストが表示されます。</p>
    <div class="url">%s</div>
    <a class="btn" href="%s">メニューに戻る</a>
  </div>
</body>
</html>`, title, title, currentURL, menuURL)
	resp.Header().Set("Content-Type", "text/html; charset=utf-8")
	resp.WriteHeader(http.StatusOK)
	_, _ = resp.Write([]byte(html))
}

func (i HeresphereResource) getHeresphereModeScenes(req *restful.Request, resp *restful.Response) {
	setHeresphereMode(req, "scenes")
	writeHeresphereReadyPage(req, resp, "従来のシーンリスト")
}

func (i HeresphereResource) getHeresphereCollection(req *restful.Request, resp *restful.Response) {
	i.writeHeresphereActorLibrary(req, resp, models.RequestActorList{
		Sort:     optional.NewString("furigana_asc"),
		MinAvail: optional.NewInt(1),
		Limit:    optional.NewInt(10000),
	}, "", "collection", true)
}

func (i HeresphereResource) getHeresphereCollectionPickup(req *restful.Request, resp *restful.Response) {
	i.writeHeresphereActorLibrary(req, resp, models.RequestActorList{
		Sort:     optional.NewString("random"),
		MinAvail: optional.NewInt(1),
		Limit:    optional.NewInt(50),
	}, "", "pickup", true)
}

func (i HeresphereResource) getHeresphereCollectionRanking(req *restful.Request, resp *restful.Response) {
	i.writeHeresphereActorLibrary(req, resp, models.RequestActorList{
		Sort:     optional.NewString("rating_desc"),
		MinAvail: optional.NewInt(1),
		Limit:    optional.NewInt(50),
	}, "総合", "ranking", true)
}

func (i HeresphereResource) getHeresphereCollectionSceneRanking(req *restful.Request, resp *restful.Response) {
	i.writeHeresphereActorLibrary(req, resp, models.RequestActorList{
		Sort:     optional.NewString("scene_rating_desc"),
		MinAvail: optional.NewInt(1),
		Limit:    optional.NewInt(50),
	}, "シーン総合", "scene-ranking", true)
}

func (i HeresphereResource) writeHeresphereActorLibrary(req *restful.Request, resp *restful.Response, query models.RequestActorList, rankPrefix string, mode string, allowReadyPage bool) {
	if !config.Config.Interfaces.DeoVR.Enabled {
		return
	}

	// ブラウザ（text/html）から直接遷移した場合はモードを保存して案内ページを返し、API切替ボタン押下時にJSONを返す
	if allowReadyPage && req.Request.Method == http.MethodGet && strings.Contains(req.HeaderParameter("Accept"), "text/html") {
		setHeresphereMode(req, mode)
		modeTitle := "コレクション：全女優リスト"
		switch mode {
		case "pickup":
			modeTitle = "ピックアップ（おすすめ女優）"
		case "ranking":
			modeTitle = "ランキング：総合"
		case "scene-ranking":
			modeTitle = "ランキング：シーン総合"
		}
		writeHeresphereReadyPage(req, resp, modeTitle)
		return
	}

	out := models.QueryActors(query, true)
	proto := getProto(req)
	host := req.Request.Host

	var actorLists []HeresphereListScenes
	rank := 0
	for _, actor := range out.Actors {
		var sceneURLs []string
		for _, scene := range actor.Scenes {
			if !scene.IsAvailable {
				continue
			}
			sceneURLs = append(sceneURLs, fmt.Sprintf("%v://%v/heresphere/%v", proto, host, scene.ID))
		}
		if len(sceneURLs) == 0 {
			continue
		}
		rank++
		name := actor.Name
		if rankPrefix != "" {
			name = fmt.Sprintf("%s%d．%s", rankPrefix, rank, actor.Name)
		}
		actorLists = append(actorLists, HeresphereListScenes{
			Name: name,
			List: sceneURLs,
		})
	}

	resp.WriteHeaderAndEntity(http.StatusOK, HeresphereLibrary{
		Access:  1,
		Library: actorLists,
	})
}

func (i HeresphereResource) getHeresphereCollection2(req *restful.Request, resp *restful.Response) {
	i.writeHeresphereCollection2Library(req, resp, true)
}

func (i HeresphereResource) writeHeresphereCollection2Library(req *restful.Request, resp *restful.Response, allowReadyPage bool) {
	if !config.Config.Interfaces.DeoVR.Enabled {
		return
	}

	if allowReadyPage && req.Request.Method == http.MethodGet && strings.Contains(req.HeaderParameter("Accept"), "text/html") {
		setHeresphereMode(req, "collection2")
		writeHeresphereReadyPage(req, resp, "コレクション2（女優＝1シーン）")
		return
	}

	proto := getProto(req)
	host := req.Request.Host

	buildActorURLList := func(query models.RequestActorList) []string {
		out := models.QueryActors(query, false)
		urls := make([]string, 0, len(out.Actors))
		for _, actor := range out.Actors {
			if actor.AvailCount <= 0 {
				continue
			}
			urls = append(urls, fmt.Sprintf("%s://%s/heresphere/actor/%d", proto, host, actor.ID))
		}
		return urls
	}

	type groupDef struct {
		name  string
		query models.RequestActorList
	}
	groups := []groupDef{
		{
			name: "全女優",
			query: models.RequestActorList{
				Sort:     optional.NewString("furigana_asc"),
				MinAvail: optional.NewInt(1),
				Limit:    optional.NewInt(10000),
			},
		},
		{
			name: "ピックアップ",
			query: models.RequestActorList{
				Sort:     optional.NewString("random"),
				MinAvail: optional.NewInt(1),
				Limit:    optional.NewInt(50),
			},
		},
		{
			name: "ランキング：総合",
			query: models.RequestActorList{
				Sort:     optional.NewString("rating_desc"),
				MinAvail: optional.NewInt(1),
				Limit:    optional.NewInt(50),
			},
		},
		{
			name: "ランキング：シーン総合",
			query: models.RequestActorList{
				Sort:     optional.NewString("scene_rating_desc"),
				MinAvail: optional.NewInt(1),
				Limit:    optional.NewInt(50),
			},
		},
	}

	library := make([]HeresphereListScenes, 0, len(groups)+10)
	for _, g := range groups {
		list := buildActorURLList(g.query)
		if len(list) == 0 {
			continue
		}
		library = append(library, HeresphereListScenes{
			Name: g.name,
			List: list,
		})
	}

	// 最近チェック
	if recentIDs := customapi.GetRecentCollectionActorIDs(); len(recentIDs) > 0 {
		db, _ := models.GetDB()
		var recentActors []models.Actor
		db.Select("id, avail_count").Where("id IN (?) AND avail_count > 0", recentIDs).Find(&recentActors)
		db.Close()

		availSet := make(map[uint]bool, len(recentActors))
		for _, a := range recentActors {
			availSet[a.ID] = true
		}
		recentURLs := make([]string, 0, len(recentIDs))
		for _, id := range recentIDs {
			if availSet[id] {
				recentURLs = append(recentURLs, fmt.Sprintf("%s://%s/heresphere/actor/%d", proto, host, id))
			}
		}
		if len(recentURLs) > 0 {
			library = append(library, HeresphereListScenes{
				Name: "最近チェック",
				List: recentURLs,
			})
		}
	}

	// ジャンルタグ別グループ
	genreGroups := []string{"美少女", "お姉さん", "癒し", "グラマー", "爆乳", "人妻", "ハーフ", "欧米", "エロ"}
	for _, tag := range genreGroups {
		list := buildActorURLList(models.RequestActorList{
			Tags:     []optional.String{optional.NewString(tag)},
			Sort:     optional.NewString("rating_desc"),
			MinAvail: optional.NewInt(1),
			Limit:    optional.NewInt(1000),
		})
		if len(list) == 0 {
			continue
		}
		library = append(library, HeresphereListScenes{
			Name: tag,
			List: list,
		})
	}

	resp.WriteHeaderAndEntity(http.StatusOK, HeresphereLibrary{
		Access:  1,
		Library: library,
	})
}

func mapHeresphereProjection(videoProjection string) (projection string, stereo string, fov float64, lens string) {
	projection = "equirectangular"
	stereo = "sbs"
	fov = 180.0
	lens = "Linear"

	switch videoProjection {
	case "flat":
		projection = "perspective"
		stereo = "mono"
	case "180_mono":
		projection = "equirectangular"
		stereo = "mono"
	case "360_mono":
		projection = "equirectangular360"
		stereo = "mono"
	case "180_sbs":
		projection = "equirectangular"
	case "360_tb":
		projection = "equirectangular360"
		stereo = "tb"
	case "mkx200":
		projection = "fisheye"
		fov = 200.0
		lens = "MKX200"
	case "mkx220":
		projection = "fisheye"
		fov = 220.0
		lens = "MKX220"
	case "vrca220":
		projection = "fisheye"
		fov = 220.0
		lens = "VRCA220"
	case "rf52", "fisheye190":
		projection = "fisheye"
		fov = 190.0
	case "fisheye":
		projection = "fisheye"
	}
	return
}

func (i HeresphereResource) getHeresphereActor(req *restful.Request, resp *restful.Response) {
	if !config.Config.Interfaces.DeoVR.Enabled {
		return
	}

	actorID, err := strconv.Atoi(req.PathParameter("actor-id"))
	if err != nil || actorID <= 0 {
		return
	}

	dnt := ""
	if !config.Config.Interfaces.DeoVR.TrackWatchTime {
		dnt = "?dnt=true"
	}

	db, _ := models.GetDB()
	defer db.Close()

	var actor models.Actor
	if err := db.Select(`actors.*, (select AVG(s.star_rating) from scene_cast sc join scenes s on s.id=sc.scene_id where sc.actor_id = actors.id and s.star_rating > 0 and is_hidden=0) as scene_rating_average`).
		Preload("Scenes", func(tx *gorm.DB) *gorm.DB {
			return tx.Where("is_available = 1 AND is_hidden = 0").Order("release_date DESC")
		}).Where("id = ?", actorID).First(&actor).Error; err != nil {
		log.Error(err)
		return
	}

	proto := getProto(req)
	host := req.Request.Host

	var media []HeresphereMedia
	var totalDuration float64
	var firstProjection string
	hasFirstProjection := false
	latestRelease := ""
	latestAdded := ""

	for _, scene := range actor.Scenes {
		if latestRelease == "" && !scene.ReleaseDate.IsZero() {
			latestRelease = scene.ReleaseDate.Format("2006-01-02")
		}
		if latestAdded == "" && !scene.CreatedAt.IsZero() {
			latestAdded = scene.CreatedAt.Format("2006-01-02")
		}

		videoFiles, err := scene.GetVideoFilesSorted(config.Config.Interfaces.Players.VideoSortSeq)
		if err != nil || len(videoFiles) == 0 {
			continue
		}

		for fileIdx, file := range videoFiles {
			if !hasFirstProjection {
				firstProjection = file.VideoProjection
				hasFirstProjection = true
			}

			resolution := strconv.Itoa(file.VideoHeight)
			if file.VideoProjection == "360_tb" {
				resolution = strconv.Itoa(file.VideoHeight / 2)
			}

			mediaName := fmt.Sprintf("%d. %s", len(media)+1, scene.Title)
			if len(videoFiles) > 1 {
				mediaName = fmt.Sprintf("%d. %s (Part %d/%d)", len(media)+1, scene.Title, fileIdx+1, len(videoFiles))
			}
			mediaName = fmt.Sprintf("%s [%vp - %s]", mediaName, resolution, humanize.Bytes(uint64(file.Size)))

			media = append(media, HeresphereMedia{
				Name: mediaName,
				Sources: []HeresphereSource{
					{
						Resolution: StringOrInt(resolution),
						Height:     file.VideoHeight,
						Width:      file.VideoWidth,
						Size:       file.Size,
						URL:        fmt.Sprintf("%s://%s/api/dms/file/%d%s", proto, host, file.ID, dnt),
					},
				},
			})
			totalDuration += file.VideoDuration
		}
	}

	thumbImg := strings.TrimSpace(actor.ImageUrl)
	if thumbImg == "" {
		thumbImg = strings.TrimSpace(actor.FaceImageUrl)
	}
	if thumbImg == "" {
		thumbImg = fmt.Sprintf("%s://%s/ui/images/blank.png", proto, host)
	} else if strings.HasPrefix(thumbImg, "/") {
		thumbImg = fmt.Sprintf("%s://%s%s", proto, host, thumbImg)
	}

	var descParts []string
	if actor.CupSize != "" {
		descParts = append(descParts, fmt.Sprintf("%sカップ", actor.CupSize))
	}
	if actor.BandSize > 0 || actor.WaistSize > 0 || actor.HipSize > 0 {
		descParts = append(descParts, fmt.Sprintf("B%d-W%d-H%d", actor.BandSize, actor.WaistSize, actor.HipSize))
	}
	if actor.Height > 0 {
		descParts = append(descParts, fmt.Sprintf("T%d", actor.Height))
	}
	descParts = append(descParts, fmt.Sprintf("動画 %d 件", len(media)))
	if actor.Biography != "" {
		descParts = append(descParts, actor.Biography)
	}

	var tags []HeresphereTag
	if actor.CupSize != "" {
		tags = append(tags, HeresphereTag{Name: "Cup:" + actor.CupSize})
	}
	if actor.GenreTags != "" {
		var genreTags []string
		if err := json.Unmarshal([]byte(actor.GenreTags), &genreTags); err == nil {
			for _, gt := range genreTags {
				if gt != "" {
					tags = append(tags, HeresphereTag{Name: gt})
				}
			}
		}
	}

	projection, stereo, fov, lens := mapHeresphereProjection(firstProjection)

	formatStar := func(val float64) string {
		if val <= 0 {
			return "-"
		}
		return strconv.FormatFloat(math.Round(val*10)/10, 'f', -1, 64)
	}
	sceneAvgVal := 0.0
	if actor.SceneRatingAverage != "" {
		if parsed, err := strconv.ParseFloat(actor.SceneRatingAverage, 64); err == nil {
			sceneAvgVal = parsed
		}
	}
	cupText := ""
	if actor.CupSize != "" {
		cupText = fmt.Sprintf(" (%sカップ)", actor.CupSize)
	}
	displayTitle := fmt.Sprintf("%s%s  総合★%s シーン総合★%s", actor.Name, cupText, formatStar(actor.StarRating), formatStar(sceneAvgVal))

	video := HeresphereVideo{
		Access:               1,
		Title:                displayTitle,
		Description:          strings.Join(descParts, " / "),
		ThumbnailImage:       thumbImg,
		DateReleased:         latestRelease,
		DateAdded:            latestAdded,
		DurationMilliseconds: uint(totalDuration * 1000),
		Rating:               actor.StarRating,
		IsFavorite:           actor.Favourite,
		Projection:           projection,
		Stereo:               stereo,
		FOV:                  fov,
		Lens:                 lens,
		Tags:                 tags,
		Media:                media,
		WriteFavorite:        false,
		WriteRating:          false,
		WriteTags:            false,
		WriteHSP:             false,
	}

	resp.WriteHeaderAndEntity(http.StatusOK, video)
}

func findTrack(indexpos int, cuepointCount int, name string, trackAssignments *[]string, scene models.Scene) (int, bool) {
	// find the track number to use for the actor, actors get their own track
	if isCast(name, scene) {
		if config.Config.Interfaces.Heresphere.MultitrackCastCuepoints {
			return getTrackAssignment(name, trackAssignments), true
		} else {
			return getTrackAssignment("cast", trackAssignments), true
		}
	}

	if !config.Config.Interfaces.Heresphere.MultitrackCuepoints {
		return -1, false
	}
	// if a position & action exist, make the action track zero, it is more likely the main cuepoint
	if cuepointCount > 1 {
		if indexpos == 0 {
			return getTrackAssignment(fmt.Sprintf("group%v", indexpos), trackAssignments), false
		}
		if indexpos == 1 {
			return 0, false
		}
		getTrackAssignment(fmt.Sprintf("group%v", indexpos), trackAssignments)
	}
	return 0, false
}

func isCast(name string, scene models.Scene) bool {
	// check if the cuepoint is the same as a cast members name
	for _, cast := range scene.Cast {
		if strings.EqualFold(cast.Name, name) {
			return true
		}
	}
	return false
}

func getTrackAssignment(name string, trackAssignments *[]string) int {
	// find tracks allocated track or add a new one
	for idx, track := range *trackAssignments {
		if strings.EqualFold(track, name) {
			return idx + 1
		}
	}
	*trackAssignments = append(*trackAssignments, name)
	return len(*trackAssignments)
}

func (fs *StringOrInt) UnmarshalJSON(data []byte) error {
	var str string
	// Try to unmarshal as a string
	if err := json.Unmarshal(data, &str); err == nil {
		*fs = StringOrInt(str)
		return nil
	}
	// Try to unmarshal as an int
	var num int
	if err := json.Unmarshal(data, &num); err == nil {
		*fs = StringOrInt(fmt.Sprintf("%d", num))
		return nil
	}
	// Return an error if neither worked
	return fmt.Errorf("invalid resolution format: %s", data)
}

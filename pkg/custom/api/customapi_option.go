package api

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"

	"github.com/xbapps/xbvr/pkg/config"
	"github.com/xbapps/xbvr/pkg/models"
	"github.com/xbapps/xbvr/pkg/tasks"
)

type GetStateResponse struct {
	CurrentState config.ObjectState  `json:"currentState"`
	Config       config.ObjectConfig `json:"config"`
	Scrapers     []models.Scraper    `json:"scrapers"`
}

type RequestSaveOptions struct {
	SaveDMMSettings	  		bool   `json:"saveDMMSettings"`
	SaveThumbnailSettings 	bool   `json:"saveThumbnailSettings"`
	DmmApiId               string `json:"dmmApiId"`
	DmmAffiliateId         string `json:"dmmAffiliateId"`
	ThumbnailEnabled       bool   `json:"thumbnailEnabled"`
	ThumbnailHourInterval  int    `json:"thumbnailHourInterval"`
	ThumbnailUseRange      bool   `json:"thumbnailUseRange"`
	ThumbnailMinuteStart   int    `json:"thumbnailMinuteStart"`
	ThumbnailHourStart     int    `json:"thumbnailHourStart"`
	ThumbnailHourEnd       int    `json:"thumbnailHourEnd"`
	ThumbnailStartDelay    int    `json:"thumbnailStartDelay"`
	ThumnailStartTime      int    `json:"thumbnailStartTime"`
	ThumbnailInterval      int    `json:"thumbnailInterval"`
	ThumbnailResolution    int    `json:"thumbnailResolution"`
	ThumbnailUseCUDAEncode bool   `json:"thumbnailUseCUDAEncode"`
}

type ConfigResource struct{}

func (i ConfigResource) WebService() *restful.WebService {
	tags := []string{"CustomOptions"}

	ws := new(restful.WebService)

	ws.Path("/api_custom/options").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/state").To(i.getState).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.POST("/save").To(i.saveOptionsTaskSchedule).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	return ws
}

func (i ConfigResource) getState(req *restful.Request, resp *restful.Response) {
	var out GetStateResponse

	tasks.UpdateState()
	out.Config = config.Config
	// out.CurrentState = config.State
	// out.Scrapers = models.GetScrapers()

	resp.WriteHeaderAndEntity(http.StatusOK, out)
}

func (i ConfigResource) saveOptionsTaskSchedule(req *restful.Request, resp *restful.Response) {
	var r RequestSaveOptions
	err := req.ReadEntity(&r)
	if err != nil {
		log.Error(err)
		return
	}

	if r.SaveDMMSettings {
		if r.DmmApiId == "" || r.DmmAffiliateId == "" {
			log.Error("DMM API ID and Affiliate ID must be set")
			resp.WriteHeaderAndEntity(http.StatusBadRequest, "DMM API ID and Affiliate ID must be set")
			return
		} else {
			config.Config.Custom.DmmAPIKey.DmmAffiliateId = r.DmmAffiliateId
			config.Config.Custom.DmmAPIKey.DmmApiId = r.DmmApiId
		}
	}
	
	if r.SaveThumbnailSettings {
		if r.ThumbnailHourEnd > 23 {
			r.ThumbnailHourEnd -= 24
		}
		config.Config.Custom.ThumbnailSchedule.Enabled = r.ThumbnailEnabled
		config.Config.Custom.ThumbnailSchedule.HourInterval = r.ThumbnailHourInterval
		config.Config.Custom.ThumbnailSchedule.UseRange = r.ThumbnailUseRange
		config.Config.Custom.ThumbnailSchedule.MinuteStart = r.ThumbnailMinuteStart
		config.Config.Custom.ThumbnailSchedule.HourStart = r.ThumbnailHourStart
		config.Config.Custom.ThumbnailSchedule.HourEnd = r.ThumbnailHourEnd
		config.Config.Custom.ThumbnailSchedule.RunAtStartDelay = r.ThumbnailStartDelay

		config.Config.Custom.ThumbnailParams.Start = r.ThumnailStartTime
		config.Config.Custom.ThumbnailParams.Interval = r.ThumbnailInterval
		config.Config.Custom.ThumbnailParams.Resolution = r.ThumbnailResolution
		config.Config.Custom.ThumbnailParams.UseCUDAEncode = r.ThumbnailUseCUDAEncode
	}

	config.SaveConfig()

	resp.WriteHeaderAndEntity(http.StatusOK, r)
}

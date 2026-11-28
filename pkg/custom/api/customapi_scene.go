package api

import (
	"encoding/json"
	"math"
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/xbapps/xbvr/pkg/models"
)

type SceneResource struct{}

func (i SceneResource) WebService() *restful.WebService {
	tags := []string{"CustomScene"}

	ws := new(restful.WebService)

	ws.Path("/api_custom/scene").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/ratings/{scene-id}").To(i.setSceneDetailedRatings).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(models.Scene{}))

	ws.Route(ws.POST("/genre_tags/{scene-id}").To(i.setSceneGenreTags).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(models.Scene{}))

	return ws
}

type RequestSetSceneDetailedRatings struct {
	RatingVideoQuality *float64 `json:"rating_video_quality"`
	RatingPlay         *float64 `json:"rating_play"`
	RatingAppeal       *float64 `json:"rating_appeal"`
}

func (i SceneResource) setSceneDetailedRatings(req *restful.Request, resp *restful.Response) {
	sceneID := req.PathParameter("scene-id")
	var r RequestSetSceneDetailedRatings
	if err := req.ReadEntity(&r); err != nil {
		log.Error(err)
		resp.WriteError(http.StatusBadRequest, err)
		return
	}

	db, _ := models.GetDB()
	defer db.Close()

	var scene models.Scene
	if err := db.Preload("Cast").Where("id = ?", sceneID).First(&scene).Error; err != nil {
		resp.WriteError(http.StatusNotFound, err)
		return
	}

	if r.RatingVideoQuality != nil {
		scene.RatingVideoQuality = math.Round(*r.RatingVideoQuality)
	}
	if r.RatingPlay != nil {
		scene.RatingPlay = math.Round(*r.RatingPlay)
	}
	if r.RatingAppeal != nil {
		scene.RatingAppeal = math.Round(*r.RatingAppeal)
	}

	// 主演女優の総合評価（複数の場合は平均点）
	castSum := 0.0
	castCount := 0
	for _, actor := range scene.Cast {
		if actor.StarRating > 0 {
			castSum += actor.StarRating
			castCount++
		}
	}
	actorRating := 0.0
	if castCount > 0 {
		actorRating = math.Round((castSum/float64(castCount))*10) / 10
	}

	// 総合評価 = 女優・画質・プレイ・抜け感のうち評価済み(>0)項目の平均
	sum := 0.0
	count := 0
	for _, v := range []float64{actorRating, scene.RatingVideoQuality, scene.RatingPlay, scene.RatingAppeal} {
		if v > 0 {
			sum += v
			count++
		}
	}
	if count > 0 {
		scene.StarRating = math.Round((sum/float64(count))*10) / 10
	} else {
		scene.StarRating = 0
	}

	if err := scene.Save(); err != nil {
		log.Error(err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteHeaderAndEntity(http.StatusOK, scene)
}

type RequestSetSceneGenreTags struct {
	GenreTags []string `json:"genre_tags"`
}

func (i SceneResource) setSceneGenreTags(req *restful.Request, resp *restful.Response) {
	sceneID := req.PathParameter("scene-id")
	var r RequestSetSceneGenreTags
	if err := req.ReadEntity(&r); err != nil {
		log.Error(err)
		resp.WriteError(http.StatusBadRequest, err)
		return
	}

	db, _ := models.GetDB()
	defer db.Close()

	var scene models.Scene
	if err := db.Where("id = ?", sceneID).First(&scene).Error; err != nil {
		resp.WriteError(http.StatusNotFound, err)
		return
	}

	tagsJSON, err := json.Marshal(r.GenreTags)
	if err != nil {
		log.Error(err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}
	scene.GenreTags = string(tagsJSON)

	if err := scene.Save(); err != nil {
		log.Error(err)
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteHeaderAndEntity(http.StatusOK, scene)
}

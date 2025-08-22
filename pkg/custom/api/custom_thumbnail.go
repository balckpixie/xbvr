package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"

	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/models"
)

type RequestThumbnailParameters struct {
	ThumnailStartTime      int    `json:"start"`
	ThumbnailInterval      int    `json:"interval"`
	ThumbnailResolution    int    `json:"resolution"`
	ThumbnailUseCUDAEncode bool   `json:"useCUDAEncode"`
}

type ThumbnailResource struct{}

func (i ThumbnailResource) WebService() *restful.WebService {
	tags := []string{"DMS"}

	ws := new(restful.WebService)

	ws.Path("/api_custom/thumbnail").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/image/{file-id}").To(i.getThumbnail).
		Param(ws.PathParameter("file-id", "File ID").DataType("int")).
		ContentEncodingEnabled(false).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.HEAD("/image/{file-id}").To(i.headThumbnail).
		Param(ws.PathParameter("file-id", "File ID").DataType("int")).
		ContentEncodingEnabled(false).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.POST("/cleanup").To(i.cleanupThumbnails).
		Metadata(restfulspec.KeyOpenAPITags, tags))
	return ws
}
func (i ThumbnailResource) getThumbnail(req *restful.Request, resp *restful.Response) {
	fileID := req.PathParameter("file-id")
	http.ServeFile(resp.ResponseWriter, req.Request, filepath.Join(common.VideoThumbnailDir, fmt.Sprintf("%v.jpg", fileID)))
}

func (i ThumbnailResource) headThumbnail(req *restful.Request, resp *restful.Response) {
	fileID := req.PathParameter("file-id")
	path := filepath.Join(common.VideoThumbnailDir, fmt.Sprintf("%v.jpg", fileID))

	// ファイルの存在とメタデータを確認
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			resp.WriteErrorString(http.StatusNotFound, "Thumbnail not found")
		} else {
			resp.WriteErrorString(http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// 空の ReadSeeker を渡すことで ServeContent を正しく動作させる
	emptyBody := bytes.NewReader([]byte{})
	http.ServeContent(resp.ResponseWriter, req.Request, filepath.Base(path), info.ModTime(), emptyBody)
}

func (i ThumbnailResource) cleanupThumbnails(req *restful.Request, resp *restful.Response) {
	var r RequestThumbnailParameters
	if err := req.ReadEntity(&r); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, "params query required")
		return
	}

	// JSONに変換
	jsonBytes, err := json.Marshal(r)
	if err != nil {
		resp.WriteErrorString(http.StatusBadRequest, "invalid params")
		return
	}
	targetParams := string(jsonBytes)

	db, _ := models.GetDB()
	defer db.Close()

	var files []models.File
	// jsonb 型の比較
	if err := db.
		Where("has_thumbnail = 1").
		Where("thumbnail_parameters::jsonb != ?::jsonb", targetParams).
		Find(&files).Error; err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, "query error")
		return
	}

	deleted := 0
	for _, f := range files {
		thumbPath := filepath.Join(common.VideoThumbnailDir, fmt.Sprintf("%d.jpg", f.ID))
		_ = os.Remove(thumbPath) // 存在しなくても無視

		// DB更新
		f.HasThumbnail = false
		f.ThumbnailParameters = json.RawMessage("{}") // 空JSONにする
		if err := db.Save(&f).Error; err == nil {
			deleted++
		}
	}

	resp.WriteHeaderAndEntity(http.StatusOK, map[string]interface{}{
		"deleted_count": deleted,
	})
}

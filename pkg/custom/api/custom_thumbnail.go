package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"os"
	"bytes"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"

	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/models"
)

type RequestThumbnailParameters struct {
	ThumnailStartTime      int    `json:"thumbnailStartTime"`
	ThumbnailInterval      int    `json:"thumbnailInterval"`
	ThumbnailResolution    int    `json:"thumbnailResolution"`
	ThumbnailUseCUDAEncode bool   `json:"thumbnailUseCUDAEncode"`
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

	ws.Route(ws.DELETE("/cleanup").To(i.cleanupThumbnails).
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
	readerr := req.ReadEntity(&r)
	if readerr != nil {
		resp.WriteErrorString(http.StatusBadRequest, "params query required")
		return
	}

	db, _ := models.GetDB()
	defer db.Close()

	// var files []models.File
	// tx := db.Model(&files)
	// tx = tx.Where("has_thumbnail = 1")
	// tx = tx.Where("thumbnail_parameters <> ?", targetParams)
	// tx.Find(&files)

	// deleted := 0
	// for _, f := range files {
	// 	thumbPath := filepath.Join(common.VideoThumbnailDir, fmt.Sprintf("%d.jpg", f.ID))
	// 	// ファイル削除（存在しなくてもエラー無視）
	// 	_ = os.Remove(thumbPath)

	// 	// DB更新
	// 	f.HasThumbnail = false
	// 	f.ThumbnailParameters = nil
	// 	if err := models.DB.Save(&f).Error; err == nil {
	// 		deleted++
	// 	}
	// }

	// resp.WriteHeaderAndEntity(http.StatusOK, map[string]interface{}{
	// 	"deleted_count": deleted,
	// })


}
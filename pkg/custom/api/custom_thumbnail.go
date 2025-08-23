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

	db, _ := models.GetDB()
	defer db.Close()
	
	// File構造体はstring型に修正されているため、直接データを読み込む
	var files []models.File
	if err := db.Where("has_thumbnail = 1").Find(&files).Error; err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, "query error")
		return
	}

	deleted := 0
	var filesToDelete []models.File

	// ターゲットとなるJSONを正規化し、比較に備える
	targetBytes, err := json.Marshal(r)
	if err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, "invalid params for comparison")
		return
	}
	var targetMap map[string]interface{}
	json.Unmarshal(targetBytes, &targetMap)

	for _, f := range files {
		var dbMap map[string]interface{}
		
		// stringとして取得したJSONデータをマップにアンマーシャル
		if err := json.Unmarshal([]byte(f.ThumbnailParameters), &dbMap); err != nil {
			// JSONが不正なレコードはスキップ
			continue
		}

		// Go側でJSONを比較するヘルパー関数
		if !jsonMapsEqual(targetMap, dbMap) {
			filesToDelete = append(filesToDelete, f)
		}
	}

	for _, f := range filesToDelete {
		thumbPath := filepath.Join(common.VideoThumbnailDir, fmt.Sprintf("%d.jpg", f.ID))
		_ = os.Remove(thumbPath) // 存在しなくても無視

		// DB更新
		f.HasThumbnail = false
		f.ThumbnailParameters = "" // string型なので空文字列にする
		if err := db.Save(&f).Error; err == nil {
			deleted++
		}
	}

	resp.WriteHeaderAndEntity(http.StatusOK, map[string]interface{}{
		"deleted_count": deleted,
	})
}

// jsonMapsEqualは、キーと値が完全に一致するかを比較するヘルパー関数です。
// JSONのキーの順序は考慮しません。
func jsonMapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if v2, ok := b[k]; !ok || !interfaceEqual(v, v2) {
			return false
		}
	}
	return true
}

// interfaceEqualは、map内のinterface{}を比較するヘルパー関数です。
func interfaceEqual(a, b interface{}) bool {
	// 型アサーションを行って、対応する型で比較
	switch a.(type) {
	case float64:
		if v, ok := b.(float64); ok {
			return a.(float64) == v
		}
	case bool:
		if v, ok := b.(bool); ok {
			return a.(bool) == v
		}
	case string:
		if v, ok := b.(string); ok {
			return a.(string) == v
		}
	}
	return false
}

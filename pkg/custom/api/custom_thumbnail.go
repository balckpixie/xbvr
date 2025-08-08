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
)

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

			// HEAD: 存在確認のみ（ボディなし）
	ws.Route(ws.HEAD("/image/{file-id}").To(i.headThumbnail).
		Param(ws.PathParameter("file-id", "File ID").DataType("int")).
		ContentEncodingEnabled(false).
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
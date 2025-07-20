package api

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/xbapps/xbvr/pkg/models"

	customcommon "github.com/xbapps/xbvr/pkg/custom/common"
)

type RequestRenameFile struct {
	FileID      uint   `json:"file_id"`
	NewFilename string `json:"filename"`
}

type RequestResetFilename struct {
	FileID  uint   `json:"file_id"`
	SceneID string `json:"scene_id"`
}

type FilesResource struct{}

func (i FilesResource) WebService() *restful.WebService {
	tags := []string{"CustomFiles"}

	ws := new(restful.WebService)

	ws.Path("/api_custom/files").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/rename").To(i.renameFile).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.POST("/resetname").To(i.resetFilename).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	return ws
}

func (i FilesResource) resetFilename(req *restful.Request, resp *restful.Response) {
	db, _ := models.GetDB()
	defer db.Close()

	var r RequestResetFilename
	err := req.ReadEntity(&r)
	if err != nil {
		log.Error(err)
		return
	}

	// Assign Scene to File
	var scene models.Scene
	err = scene.GetIfExist(r.SceneID)
	if err != nil {
		log.Error(err)
		return
	}

	var f models.File
	err = db.Preload("Volume").Where(&models.File{ID: r.FileID}).First(&f).Error
	if err == nil {
		f.SceneID = scene.ID
		f.Save()
	}

	//---------------
	scene = customcommon.RenameFileBySceneID(scene, f)
	resp.WriteHeaderAndEntity(http.StatusOK, scene)
}

func (i FilesResource) renameFile(req *restful.Request, resp *restful.Response) {
	var r RequestRenameFile
	err := req.ReadEntity(&r)
	if err != nil {
		log.Error(err)
		return
	}
	scene := customcommon.RenameFileByFileId(uint(r.FileID), "", r.NewFilename)
	resp.WriteHeaderAndEntity(http.StatusOK, scene)
}

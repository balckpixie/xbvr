package api

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"

	customtasks "github.com/xbapps/xbvr/pkg/custom/tasks"
)

type TaskResource struct{}

func (i TaskResource) WebService() *restful.WebService {
	tags := []string{"CustomTask"}

	ws := new(restful.WebService)

	ws.Path("/api_custom/task").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/preview/generate").To(i.thumbnailGenerate).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	return ws
}

func (i TaskResource) thumbnailGenerate(req *restful.Request, resp *restful.Response) {
	go customtasks.GenerateThumnbnails(nil)
}

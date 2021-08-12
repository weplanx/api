package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kainonly/go-bit/crud"
	"go.uber.org/fx"
	"lab-api/controller"
)

type Dependency struct {
	fx.In

	*controller.Index
	*controller.Resource
	*controller.Admin
}

func Initialize(r *gin.Engine, d Dependency) {
	r.GET("/", crud.Bind(d.Index.Index))
	resourceRoute := r.Group("resource")
	{
		resource := d.Resource
		resourceRoute.POST("originLists", crud.Bind(resource.OriginLists))
		resourceRoute.POST("lists", crud.Bind(resource.Lists))
		resourceRoute.POST("get", crud.Bind(resource.Get))
		resourceRoute.POST("add", crud.Bind(resource.Add))
		resourceRoute.POST("edit", crud.Bind(resource.Edit))
		resourceRoute.POST("delete", crud.Bind(resource.Delete))
	}
}

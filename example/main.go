package main

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"swag/example/api"
	"app/docs"
)

// @title GOLANG-GIN
// @version 1.0
// @description 后台管理模块
// @contact.email mrjnamei@gmail.com
// @BasePath /v1
// @tags users asdasdasd
// @tags contents asdasdasdasd
func main()  {
	g := gin.New()
	g.GET("/", api.Api)
	g.GET("/api-docs", func(context *gin.Context) {
		context.String(200, docs.ReadDoc())
	})
	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

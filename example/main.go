package main

import (
	"github.com/gin-gonic/gin"
)

// @title GOLANG-GIN
// @version {1.2.1}
// @description 后台管理模块
// @contact.email  mrjnamei@gmail.com
// @BasePath /v1
// @tags common 公共部分
// @tags contents 内容部分
func main()  {
	g := gin.Default()
	g.POST("/v1/pet")
}

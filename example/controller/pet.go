package controller

import "github.com/gin-gonic/gin"

// @Summary getPets
// @Description 获取pets
// @ID file.upload
// @Accept  json
// @Produce  json
// @tag users
// @Param   page query string false  "page of the gets"
// @Success 200 {object} @Pets  "petslist"
// @Router /pets [get]
func GetPets(ctx *gin.Context)  {
	//
}

// @Summary getPets
// @Description 获取pets
// @ID file.upload
// @Accept  json
// @Produce  json
// @tag users
// @Param   pets body @Pets true "pets fields"
// @Success 200 {object} @Pets  "success"
// @Failure 422 {object} @Error  "error info"
// @Router /pets [get]
func CreatePets(ctx *gin.Context)  {
	//
}

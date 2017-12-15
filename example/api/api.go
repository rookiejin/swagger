package api

import "github.com/gin-gonic/gin"


// @Summary asd
// @Description asdasdasdasd
// @ID file.upload
// @Accept  multipart/form-data
// @Produce  json
// @tag common
// @Param   file formData file true  "sss"
// @Success 200 {string} string model.APISuccess "ok"
// @Failure 400 {object} model.File "We need ID!!"
// @Failure 404 {object} model.APIError "Can not find ID"
// @Router /file/upload [post]
func Api(ctx *gin.Context)  {
	// todo
}

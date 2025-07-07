package middlerware

import (
	"GIN/config"
	"GIN/model"
	"GIN/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)
func DeactivateUser()gin.HandlerFunc{
return func(c *gin.Context){
	if err:=config.DB.Model(&model.Users{}).Where("id=?",c.GetInt("id")).Update("is_active",false).Error;err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	return	
}
}
}
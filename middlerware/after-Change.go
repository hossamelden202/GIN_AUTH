package middlerware

import (
	"GIN/config"
	"GIN/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)
func AfterChange() gin.HandlerFunc{
return func(c *gin.Context){
if err:=config.Rdb.FlushAll(config.Ctx).Err();err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	return
}
c.SetCookie("refresh_token","1",-1,"/","localhost",false,true)

c.Next()
}
}

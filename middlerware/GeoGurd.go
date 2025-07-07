package middlerware

import (
	"GIN/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)
func GeoGurd()gin.HandlerFunc{
return func(c *gin.Context){
	data,err:=utils.Sendlocation(c.Request.RemoteAddr)
	if err!=nil{
		utils.SendError(c,http.StatusInternalServerError,"something went wrong")
		return
	}
	if data.Country=="Ukraine"||data.Country=="Russia"{
	
		utils.SendError(c,http.StatusUnauthorized,"users in this country cannot access my website ,fuck you")
		return
	}
}


}
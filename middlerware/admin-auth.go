package middlerware

import (
	"GIN/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)
func AdminAuth()gin.HandlerFunc{
	return func(c *gin.Context){

		role,exist:=c.Get("role")
		if !exist||role!="admin" {
			utils.SendError(c,http.StatusUnauthorized,"only admin can enter")
			return
		}
	}
}
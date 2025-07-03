package middlerware

import (
	"GIN/config"
	"net/http"
	

	"github.com/gin-gonic/gin"
)

func Reauth()gin.HandlerFunc{
	return func(c *gin.Context){
	
	email,exist:=c.Get("email")
		if !exist{
		c.AbortWithStatusJSON(http.StatusUnauthorized,"you didnot authorize your action choose authorization method")
		return
	}
	emailstr,ok:=email.(string)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError,"something went wrong")
		return
	}

	num,err:=config.Rdb.Exists(config.Ctx,"Reauth:"+emailstr).Result()
	if err:=config.Rdb.Del(config.Ctx,"Reauth:"+emailstr).Err();err!=nil{
		c.AbortWithStatusJSON(http.StatusInternalServerError,"something went wrong")
		return
	}
	if err!=nil||num==0{
		c.AbortWithStatusJSON(http.StatusUnauthorized,"you didnot authorize your action choose authorization method")
		return
	}
	 c.Next()


	}
}

package middlerware

import (
	"GIN/config"
	"GIN/utils"

	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)
func ModeratorAuth()gin.HandlerFunc{
	return func(c *gin.Context){
		header:=c.GetHeader("Authorization")
		if header==""{
			utils.SendError(c,http.StatusUnauthorized,"no Authorziation header")
			return
		}
		
		secret:=[]byte(os.Getenv("jwt_secret"))
		e,exist:=strings.CutPrefix(header,"Bearer")
		if !exist{
			utils.SendError(c,http.StatusUnauthorized,"no valid token exist")
			return
		}
		TokenString:=strings.ReplaceAll(e," ","")
		token,err:=jwt.Parse(TokenString,func(token *jwt.Token)(interface{},error){

			if _,ok:=token.Method.(*jwt.SigningMethodHMAC);!ok{
				utils.SendError(c,http.StatusUnauthorized,"in vaild token been tempared with")
				return nil,jwt.ErrSignatureInvalid
			}	
			return secret,nil

		})
		if err==nil||!token.Valid{
			utils.SendError(c,http.StatusUnauthorized,"invlaid token signuture")
			return
		}
		if Claims,ok:=token.Claims.(jwt.MapClaims);ok{
					if jti,ok:=Claims["jti"];ok{
		if result,err:=config.Rdb.Exists(config.Ctx,"Blocklist:"+jti.(string)).Result();err!=nil||result==1{
			utils.SendError(c,http.StatusUnauthorized,"you already logged out siginin again")
			return
		}

		}
 
	if exp,ok:=Claims["exp"].(float64); ok{
		if time.Now().Unix()>int64(exp){
			utils.SendError(c,http.StatusUnauthorized,"exp time is out")
			return
		}
	}
if id,ok:=Claims["id"].(float64);ok{
	c.Set("id",int(id))
}
if Username,ok:=Claims["Username"];ok{
	c.Set("Username",Username)
}
if email,ok:=Claims["email"];ok{
	c.Set("email",email)
}
if role,ok:=Claims["role"];ok{
	c.Set("role",role)
}

}
 role,exist:=c.Get("role")
 if !exist ||role!="moderator"{
	utils.SendError(c,http.StatusUnauthorized,"only moderator can enter this section")
 }

c.Next()
	}
}
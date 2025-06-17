package middlerware

import (
	"GIN/utils"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	//"golang.org/x/text/internal"
)
func Auth() gin.HandlerFunc{
return func(c *gin.Context){
secret:=[]byte(os.Getenv("jwt_secret"))
header:=c.GetHeader("Authorization")
token2,flag:=strings.CutPrefix(header,"bearer")
if !flag{
	utils.SendError("invaild token")
	return
}
token:=strings.ReplaceAll(token2," ","")
tokent,err:=jwt.ParseWithClaims(token,func(token *jwt.Token)(interface{},error){
if _ok:=tokent.Method.(*jwt.SigningMethod);ok!=nil{
	return nil,jwt.ErrSignatureInvalid
}
return secret,nil
})




}
}

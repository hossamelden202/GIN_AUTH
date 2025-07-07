package middlerware

import (
	"GIN/config"
	"GIN/model"
	"GIN/utils"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	//"golang.org/x/text/internal"
)
func Auth() gin.HandlerFunc{
return func(c *gin.Context){
secret:=[]byte(os.Getenv("jwt_secret"))
header:=c.GetHeader("Authorization")
token2,flag:=strings.CutPrefix(header,"Bearer")
if !flag{
	utils.SendError(c,http.StatusUnauthorized,"invaild token")
	return
}
tokenstring:=strings.ReplaceAll(token2," ","")

token,err:=jwt.Parse(tokenstring,func(token *jwt.Token) (interface{},error){
if _,ok:=token.Method.(*jwt.SigningMethodHMAC); !ok{

return nil,jwt.ErrSignatureInvalid
}
return secret,nil

})

if err!=nil||!token.Valid{
	utils.SendError(c,http.StatusUnauthorized,"unauthorized method")
	fmt.Println(token)
	fmt.Println("validation:",token.Valid)
	return
}
 var devid float64
 var devinfo model.DeviceRecord
if Claims,ok:=token.Claims.(jwt.MapClaims);ok{
var version interface{}
var exist bool
	if version,exist=Claims["version"];!exist{
		utils.SendError(c,http.StatusUnauthorized,"you are unauthorized")
		return
	}else{
		c.Set("version",version)
	}
	if email,ok:=Claims["email"];ok{
	var user model.Users
if  err:=config.DB.Model(&model.Users{}).Where("email=?",email).Find(&user).Error;err!=nil{	
	utils.SendError(c,http.StatusUnauthorized,"something went wrong")
	return
}
if user.Token_version!=int(version.(float64)){
utils.SendError(c,http.StatusUnauthorized,"token version is old try logging in again")
return
}
		c.Set("email",email)

}else {
	return
}	
	if jti,ok:=Claims["jti"];ok{
		if result,err:=config.Rdb.Exists(config.Ctx,"Blocklist:"+jti.(string)).Result();err!=nil||result==1{
			utils.SendError(c,http.StatusUnauthorized,"you already logged out siginin again")
			return
		}
		c.Set("jti",jti.(string))
fmt.Println("⚠️  jti is found equal",jti)
		}else{
			fmt.Println("⚠️  jti missing from JWT")
		}
	if exp,ok:=Claims["exp"].(float64); ok{
		if time.Now().Unix()>int64(exp){
			utils.SendError(c,http.StatusUnauthorized,"exp time is out")
			return
		}
	}else{
		fmt.Println("⚠️  exp missing from JWT")
	}
if id,ok:=Claims["id"].(float64);ok{
	c.Set("id",int(id))
	fmt.Println("⚠️  id is found equal",id)
}else{
	fmt.Println("⚠️  id missing from JWT")
}
if Username,ok:=Claims["Username"];ok{
	c.Set("Username",Username)
	fmt.Println("⚠️  username is found equal",Username)
}else{
	fmt.Println("⚠️  username missing from JWT")
}

if role,ok:=Claims["role"];ok{
	c.Set("role",role)
}else{
	fmt.Println("⚠️  role missing from JWT")
}
if devid,ok=Claims["devid"].(float64);ok{
	c.Set("devid",int(devid))

if err:=config.DB.Model(&model.DeviceRecord{}).Where("id=?",int(devid)).First(&devinfo).Error;err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	return
}
}

}

date,err:=utils.Sendlocation(c.Request.RemoteAddr)
if err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	return
}
if date.City!=devinfo.City{
	utils.SendError(c,http.StatusUnauthorized,"your city that was logged has changed u must log in again")
	return
}



c.Next()
}

}

package utils

import (
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	 "golang.org/x/crypto/bcrypt"

)

func GenerteJwt(Username string,Email string ,id int,role string)(string,error){
token:=jwt.NewWithClaims(jwt.SigningMethodHS256,jwt.MapClaims{
	"Username":Username,
	"email":Email,
	"role":role,
	"id":id,
	"exp":time.Now().Add(time.Hour*24).Unix(),
	"iat":time.Now().Unix(),


})


return token.SignedString([]byte(os.Getenv("jwt_secret")))

}
func GenterUserName(name string)string{
sub1:=uuid.New().String()

sub:=sub1[:8]
strings.ToLower(name)
strings.ReplaceAll(name," ","")
re:=regexp.MustCompile(`[^a-zA-Z]+`)
clean := re.ReplaceAllString(name, "")
username:=clean+"_"+ string(sub)
return username
}
func ParseANDSendResponse(c *gin.Context,input interface{})bool{
switch c.ContentType(){
case "application/xml":
if err:=c.ShouldBindXML(input); err!=nil{
	c.XML(http.StatusBadRequest,gin.H{"error":err.Error()})
	return false
}
case "application/x-yaml":
	if err:=c.ShouldBindBodyWithYAML(input);err!=nil{
		c.YAML(http.StatusBadRequest,gin.H{"error":err.Error()})
return false
	}

default:
	if err:=c.ShouldBindJSON(input);err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
		return false
	}

}
return true
}
func SendRes(c *gin.Context,res interface{}){
	switch c.ContentType(){
case "application/xml":

	c.XML(http.StatusOK,res)


case "application/x-yaml":
c.YAML(http.StatusOK,res)
default:
	c.JSON(http.StatusOK,res)
}

}
func SendError(c *gin.Context,status int,err string){

	switch c.ContentType(){
		case "application/xml":
			if err!=""{
			c.XML(status,gin.H{"error":err})
			return
			}
		case "application/x-yaml":
			if err!=""{
			c.YAML(status,gin.H{"error":err})
			return
			}
		default:
			if err!=""{
				c.JSON(status,gin.H{"error":err})
			}
	}
}
func HashPassword(c *gin.Context,password string)string{
bytes,err:=bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)
if err!=nil{
SendError(c,http.StatusForbidden,err.Error())
return "err"
}
return string(bytes)

}
func CheckEMail(email string) bool{
if !strings.Contains(email,"@"){
	return false
}

str:= strings.Split(email,"@")
   if len(str) != 2 {
    return false
  }

if str[1]!="gmail.com"{
return false}

valid:=regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
if !valid.MatchString(str[0]){
return false
}

return true
}
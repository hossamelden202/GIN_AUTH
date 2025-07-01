package controllers

import (
	//"net/http"

	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	// "github.com/gin-gonic/gin/binding"
	"GIN/config"
	"GIN/model"
	"GIN/utils"
)

// {
//   "email":"hossam@gmail.com",
  
//   "password":"ue2dh3"
 
// }
type input struct{
 Name string `json:"name" binding:"required,min=2,max=255,alpha_space"`
 Email string `json:"email" binding:"required,email"`
Country string `json:"Country" binding:"required"`
ZipCode string `json:"ZipCode" binding:"required"`
Phone string `json:"phone" binding:"required,numeric"`
Password string `json:"password" binding:"required"`
Gender string `json:"gender" binding:"required,oneof=male female"`

}
type login struct{
 Email string `json:"email" binding:"required,email"`
Password string `json:"password" binding:"required"`
}
// CREATE TABLE AddressUser (

//  id SERIAL PRIMARY KEY
//  ,userID  INT REFERENCES users(id) ON DELETE CASCADE
//  ,Street VARCHAR(500)
//  ,City VARCHAR(500)
//  ,State VARCHAR(100)
// ,Country VARCHAR(500),
//  locale VARCHAR(10)
//  ,ZipCode VARCHAR(100) ,
//  timezone VARCHAR(100)
// );
// CREATE TABLE users(
//     id SERIAL PRIMARY KEY,
//     name VARCHAR(255) NOT NULL
//     ,usename VARCHAR(100) NOT NULL UNIQUE,
//     email VARCHAR(320) NOT NULL UNIQUE,
//     isEmailVerified BOOLEAN default false ,
//     verificationCode VARCHAR(255),
//     verificationExpiresAt TIMESTAMP 

// ,   phone 	VARCHAR(20) UNIQUE,

//     passwordHash VARCHAR(255)

// ,    role VARCHAR(10) CHECK (role IN ('admin','user','moderator','staff'))
 
//  ,profileImageUrl VARCHAR(500),
//  coverImageUrl TEXT
//  ,bio TEXT,
//  isVerified BOOLEAN,
//  gender VARCHAR CHECK (gender IN ('male','female')),


//  birthday DATE,
//  isActive BOOLEAN DEFAULT true,
//  lastLoginAt TIMESTAMP WITH TIME ZONE,
//  createdAt TIMESTAMP WITH TIME ZONE,
//  updatedAt TIMESTAMP WITH TIME ZONE,
//  deletedAt TIMESTAMP WITH TIME ZONE
 
// );
func Create(c *gin.Context){
var input input
if utils.ParseANDSendResponse(c,&input)==false {
return
}
if !utils.CheckEMail(input.Email){
	utils.SendError(c,http.StatusBadRequest,"invaild email")
	return
}
	username:=utils.GenterUserName(input.Name)
var hashed string
	if hashed=utils.HashPassword(c,input.Password);hashed=="err"{
		return
	}
	users:=model.Users{
Username :username,
Name:input.Name,
Email: input.Email,
Phone: input.Phone,
Role:"user",
PasswordHash: hashed,
Gender: input.Gender,
	}
errD:=config.DB.Create(&users).Error
if errD!=nil{
utils.SendError(c,http.StatusBadRequest,errD.Error())
fmt.Println("something went wrong",errD.Error())
return}
// 	Token,errT:=utils.GenerteJwt(username,input.Email,int(users.ID),users.Role,time.Minute*7)
// 	if errT!=nil{
// 	utils.SendError(c,http.StatusNotAcceptable,errT.Error())
// return}
	

res:=struct{
User model.Users `json:"user"`

}{
User:users,

}
utils.SendRes(c,res)
}



func Login(c *gin.Context){
	var input login
	utils.ParseANDSendResponse(c,&input)
if !utils.CheckEMail(input.Email){
	utils.SendError(c,http.StatusForbidden,"enter valid email")
	return
}

if utils.Attempts(c,input.Email)==true{
return
}
var user model.Users

if res:=config.DB.Where("email = ?",input.Email).First(&user);!utils.CheckPass(input.Password,user.PasswordHash)|| res.RowsAffected==0||res.Error!=nil {
	 if utils.IncrAttempts(c,input.Email)==false{
		return
	 }
	utils.SendError(c,http.StatusUnauthorized,"try again and enter your info correctly")
	return
}

acctoken,err:=utils.GenerteJwt(user.Username,user.Email,int(user.ID),user.Role,time.Minute*15)

if err!=nil{
	utils.SendError(c,http.StatusBadRequest,"try again someting went wrong")
	return
}
RefreshToken,errR:=utils.GenerteJwt(user.Username,user.Email,int(user.ID),user.Role,time.Hour*7*24)

if errR!=nil{
	utils.SendError(c,http.StatusBadRequest,"try again someting went wrong")
	return

}
c.SetCookie("refresh_token",RefreshToken,24*60*7*60,"/","localhost",false,true)

response:=struct{
	User model.Users
	Token string
	Refresh_token string
}{
	User:user,
	Token:acctoken,
	Refresh_token: RefreshToken,
}
utils.SendRes(c,response)
}



func Refresh(c *gin.Context){
RefreshToken,err:=c.Cookie("refresh_token")
if err!=nil{
	utils.SendError(c,http.StatusUnauthorized,"no refreshToken cookie has been set cannot renew seasion")
	return
}
if _,ok:=utils.VaildateToken(c,RefreshToken);!ok{
utils.SendError(c,http.StatusUnauthorized,"no valid refreshToken cookie has been set cannot renew seasion")
	return
} 
var username,email,role string
var id int
if val,exist:=c.Get("Username");exist{
username=val.(string)
}
if val,exist:=c.Get("email");exist{
	email=val.(string)
}
if val,exist:=c.Get("id");exist{
	id=val.(int)
	fmt.Println("here is id ",id)
}
if val,exist:=c.Get("role");exist{
	role=val.(string)
}
 newToken,errt:=utils.GenerteJwt(username,email,id,role,time.Minute*7)
 if errt!=nil{
	utils.SendError(c,http.StatusBadRequest,"something went wrong try again")
	return
 }
 res:=struct{
	NewAcesstoken string
 }{
	NewAcesstoken:newToken,
 }
 utils.SendRes(c,res)
}
func Logout(c *gin.Context){
var refresh_Token string
var err error
	if refresh_Token,err=c.Cookie("refresh_token");err!=nil{
		utils.SendError(c,http.StatusUnauthorized,"refresh token expried u already logged out")
		fmt.Println(err.Error())
		return
	}
	refreshToken,ok1:=utils.VaildateToken(c,refresh_Token)
	
	if !ok1{
		utils.SendError(c,http.StatusUnauthorized,"refreshToken invalid")
		return
	}


Header:=c.GetHeader("Authorization")
if Header==""{
	utils.SendError(c,http.StatusUnauthorized,"No Authorization Header")
	return
}
var tokenString string
var ok bool
	if tokenString,ok=strings.CutPrefix(Header,"Bearer");!ok{
		utils.SendError(c,http.StatusUnauthorized,"Not valid token should start with Bearer")
		return
	}
	tokenString=strings.TrimSpace(tokenString)
	if strings.HasPrefix(tokenString,"&{"){
	if tokenString,ok=strings.CutPrefix(tokenString,"&{");!ok{
		utils.SendError(c,http.StatusUnauthorized,"something went wrong cleaning token")
		return
	}
	}

 
secret:=[]byte(os.Getenv("jwt_secret"))

 token,okay:=jwt.Parse(tokenString,func(token *jwt.Token)(interface{},error){
	if _,ok:=token.Method.(*jwt.SigningMethodHMAC);!ok{
		return nil,jwt.ErrSignatureInvalid
	}
	return secret,nil
})
if  okay!=nil||!token.Valid{
	utils.SendError(c,http.StatusUnauthorized,"Not valid token signuture")
	fmt.Println(okay.Error())
	fmt.Println(token)
	return
}
if refreshToken==token{
	utils.SendError(c,http.StatusUnauthorized,"token and refreshToken cannot be the same")
	return
}
c.SetCookie("refresh-token","",-1,"/",os.Getenv("domain"),true,true)
if cla,ok:=token.Claims.(jwt.MapClaims);ok{
	var exp interface{}
	var ok bool
	if exp,ok=cla["exp"];!ok{
		utils.SendError(c,http.StatusFailedDependency,"something went wrong")
		return
	}
	if jti,ok:=cla["jti"];ok{
		ttl:=time.Until(time.Unix(exp.(int64),0))
		config.Rdb.Set(config.Ctx,"Blocklist:"+jti.(string),"1",ttl)
		c.Set("jti",jti)
	}
	if id,ok:=cla["id"];ok{
		id:=id.(float64)
		c.Set("id",int(id))
	}
}

}
func Captcha(c *gin.Context){

	email:=c.Query("email")
	if email==""||!utils.CheckEMail(email){
		utils.SendError(c,http.StatusBadRequest,"Email is Required")
		return
	}
	err:=config.Rdb.Set(config.Ctx,"captcha:passed:"+email,"1",15*time.Minute).Err()
	if err!=nil{
		utils.SendError(c,http.StatusInternalServerError,"something went wrong")
		return
	}
	res:=struct{
	
		Message string
	}{
	
		Message:"You passed the Captcha you can login now ",
	}
	utils.SendRes(c,res)
}
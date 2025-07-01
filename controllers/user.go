package controllers

import (
	//"net/http"

	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"

	"math/big"
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
// type TokenInput struct{
// 	Username string `json:"username"`
//     Role     string `json:"role"`
//     ID       int    `json:"id"`
// }
type Verify struct {
	Code string `json:"code" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}
type Email struct{
Email string `json:"email" binding:"required,email"`
}
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
num,err:=rand.Int(rand.Reader,big.NewInt(1000000))
if err!=nil{utils.SendError(c,http.StatusInternalServerError,"something went wrong")}
err2:=config.Rdb.Set(config.Ctx,"Signup:code:"+input.Email,fmt.Sprintf("%06d",num.Int64()),15*time.Minute).Err()

if err2!=nil{utils.SendError(c,http.StatusInternalServerError,"something went wrong")}

message:=fmt.Sprintf(
`Subject: Verify Your Email Address 🚀

Hello %s,

Thank you for signing up to SOAH! 🎉  
To complete your registration and verify your email address, please enter the verification code below:

🔐 Your verification code: %s

This code will expire in 15 minutes. ⏳  
If you didn’t sign up for this account, you can safely ignore this message.

Welcome aboard,  
— The SOAH Security Team 🛡️
`,input.Name,fmt.Sprintf("%06d",num.Int64()))
utils.SendEmailSmtp(c,input.Email,message)
if con
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
utils.SendEmail(c,users.Email)
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

bytes,err:=json.Marshal(user)
if err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	return
}
err = config.Rdb.Set(config.Ctx, "Login:user:"+input.Email, bytes, 15*time.Minute).Err()
if err!=nil{
   utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	return
}
n,err:=rand.Int(rand.Reader ,big.NewInt(100000))
if err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	return
}
err2:=config.Rdb.Set(config.Ctx,"Login:code:"+input.Email,fmt.Sprintf("%06d",n.Int64()),15*time.Minute).Err()

if err2!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	fmt.Println("errors in seting ",err2.Error())
	return
}
Message:=fmt.Sprintf(`
Hello %s,

🎉 Your login was successful!

As an added layer of protection, please verify your identity using the code below:

🔐 Your 2FA verification code: %s

This code is valid for 10 minutes. Please do not share it with anyone.

If this wasn’t you, we recommend immediately changing your password.

Stay secure,  
Your Security Team 🛡️

`,input.Email,fmt.Sprintf("%06s",n.String()))
utils.SendEmailSmtp(c,input.Email,Message)
if num,err:=config.Rdb.Exists(config.Ctx,"Login:verified:"+input.Email).Result();err!=nil||num==0{
	res:=struct{
		Message string
	}{
		Message:"you must enter verification code to continue" ,
	}
	utils.SendRes(c,res)
return
}

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
    func Verify_email(c *gin.Context){
	var input Verify
	
	if !utils.ParseANDSendResponse(c,&input){
    return
	}
	code,err1:=config.Rdb.Get(config.Ctx,"Login:code:"+input.Email).Result()
	if err1!=nil{
		utils.SendError(c,http.StatusInternalServerError,"something went wrong")
		fmt.Println("we",err1.Error())
		return
	}
fmt.Println("here is code",input.Code,"sec",code)
	if input.Code==code{
		err2:=config.Rdb.Del(config.Ctx,"Login:code:"+input.Email).Err()
		if err2!=nil{
				utils.SendError(c,http.StatusInternalServerError,err2.Error())
				return
		}
	errors:=config.Rdb.Set(config.Ctx,"Login:verified:"+input.Email,"1",time.Minute*10).Err()
		if errors!=nil{
				utils.SendError(c,http.StatusInternalServerError,"something went wrong")
				// fmt.Println("why",errors)
				return
		}
		
	if config.DB.Model(&model.Users{}).Where("email=?&&tfa_verifed=?",input.Email,true).RowsAffected!=0{
utils.SendRes(c,"Enter your 2FA code to login")
return
}else{
	bytes,err3:=config.Rdb.Get(config.Ctx,"Login:user:"+input.Email).Result()
	if err3!=nil{
		utils.SendError(c,http.StatusInternalServerError,"something went wrong")
		// fmt.Println("weee",err3)
		return
	}
var user model.Users
errr:=json.Unmarshal([]byte(bytes),&user)
	if errr!=nil{
		utils.SendError(c,http.StatusInternalServerError,"something went wrong")
		fmt.Println(errr)
		return
	}
	acctoken,err4:=utils.GenerteJwt(user.Username,input.Email,int(user.ID),user.Role,time.Minute*15)

if err4!=nil{
	utils.SendError(c,http.StatusBadRequest,"try again someting went wrong")
	return
}
RefreshToken,errR:=utils.GenerteJwt(user.Username,input.Email,int(user.ID),user.Role,time.Hour*7*24)

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
utils.RestAttempts(input.Email)
utils.SendRes(c,response)
}
	

}else{
		utils.SendError(c,http.StatusUnauthorized,"Enter Email verification code correctly")
return
	}
}
func Create_2fA(c *gin.Context){
	var input Email
	utils.ParseANDSendResponse(c,&input)
	secret,err:=totp.Generate(totp.GenerateOpts{Issuer: "SOAH",AccountName: input.Email})

if err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	fmt.Println("we",err)
	return
}
error:=config.Rdb.Set(config.Ctx,"Login:2fa:"+input.Email,secret.Secret(),10*time.Minute).Err()
if error!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	fmt.Println("wqd",error)
	return
}
// content:=fmt.Sprintf("otpauth://totp/SOAH:%s?secret=%s&issuer=%s",input.userName,secret,"SOAH")
png,err:=qrcode.Encode(secret.URL(),qrcode.Medium,256)
if err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	fmt.Println("sdfa",err)
	return
}
str:=base64.StdEncoding.EncodeToString(png)
res:=struct {
	Email string
	Png string
	Secret string
   }{
	Email:input.Email,
	Png:"data:image/png;base64," + str,
	Secret:  secret.Secret(),

   }
utils.SendRes(c,res)
err5:=config.DB.Model(&model.Users{}).Where("email=?",input.Email).Update("tfa_verifed",true).Error
if err5!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	return
}
}

func Verify_2fA(c *gin.Context){
	var input Verify
	if !utils.ParseANDSendResponse(c,&input){
return
	}
	secret,err:=config.Rdb.Get(config.Ctx,"Login:2fa:"+input.Email).Result()
	if err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	fmt.Println("we",err.Error())

	return
}

	if !totp.Validate(input.Code,secret){
		utils.SendError(c,http.StatusUnauthorized,"Enter valid 2fA code")
		fmt.Println("wewww")
		return
	}else{
bytes,err3:=config.Rdb.Get(config.Ctx,"Login:user:"+input.Email).Result()
	if err3!=nil{
		utils.SendError(c,http.StatusInternalServerError,"something went wrong")
		fmt.Println("wee",err3.Error())
		return
	}
var user model.Users
errr:=json.Unmarshal([]byte(bytes),&user)
	if errr!=nil{
		utils.SendError(c,http.StatusInternalServerError,"something went wrong")
		return
	}
	acctoken,err4:=utils.GenerteJwt(user.Username,input.Email,int(user.ID),user.Role,time.Minute*15)

if err4!=nil{
	utils.SendError(c,http.StatusBadRequest,"try again someting went wrong")
	return
}
RefreshToken,errR:=utils.GenerteJwt(user.Username,input.Email,int(user.ID),user.Role,time.Hour*7*24)

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
utils.RestAttempts(input.Email)
utils.SendRes(c,response)
    

	}
}
package controllers

import (
	//"net/http"

	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"


	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"
	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"

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
type Password struct{
Password string  `json:"password" binding:"required"`
Confirm string `json:"confirm" binding:"required"`

}
type DeviceRecordStruct struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	Street    string `json:"street"`
	City      string `json:"city"`
	State     string `json:"state"`
	Country   string `json:"country"`
	Locale    string `json:"locale"`
	ZipCode   string `json:"zip_code"`
	LastLogin string `json:"last_login"`
	Browser   string `json:"browser"`
}

type Reset struct{
		Token string `json:"token" binding:"required"`
	Email string `json:"email" binding:"required,email"`
Password string `json:"password" binding:"required"`
}
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
if !utils.ParseANDSendResponse(c,&input) {
return
}
if res:=utils.ValidatePassword(input.Password);res!="0"{
	utils.SendError(c,http.StatusBadRequest,res)
}

if !utils.CheckEMail(input.Email){
	utils.SendError(c,http.StatusBadRequest,"invaild email")
	return
}
num,err:=rand.Int(rand.Reader,big.NewInt(1000000))
if err!=nil{utils.SendError(c,http.StatusInternalServerError,"something went wrong")
return}
err2:=config.Rdb.Set(config.Ctx,"Signup:code:"+input.Email,fmt.Sprintf("%06d",num.Int64()),15*time.Minute).Err()

if err2!=nil{utils.SendError(c,http.StatusInternalServerError,"something went wrong")
return}

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
byte,err:=json.Marshal(input)
if err!=nil{utils.SendError(c,http.StatusInternalServerError,"something went wrong")
return}
config.Rdb.Set(config.Ctx,"Signup:user:"+input.Email,byte,15*time.Minute)
if config.DB.Model(&model.Users{}).Where("email=?&&is_email_verified",input.Email,true).RowsAffected==0{
utils.SendRes(c,"Verify your email to continue")
return
}
	
}
func Verify_signup_email(c *gin.Context){
	var inp Verify
	if !utils.ParseANDSendResponse(c,&inp){
		return
	}
code,err:=config.Rdb.Get(config.Ctx,"Signup:code:"+inp.Email).Result()
if err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	fmt.Println(err.Error())
	return
}
if subtle.ConstantTimeCompare([]byte(inp.Code),[]byte(code))==1{
if err:=config.DB.Model(&model.Users{}).Where("email=?",inp.Email).Update("is_email_verified",true).Error;err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
		fmt.Println(err.Error())
	return

}
user,err:=config.Rdb.Get(config.Ctx,"Signup:user:"+inp.Email).Result()
if err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
		fmt.Println(err.Error())
	return
}
var input input
erre:=json.Unmarshal([]byte(user),&input)
if erre!=nil{
		utils.SendError(c,http.StatusInternalServerError,"something went wrong")
		fmt.Println(erre.Error())
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
fmt.Println(errD)
return
}
if !utils.AddpasswordHistory(hashed,users.ID){
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	return
}
// config.DB.Model(&model.Users{}).Where("email=?",input.Email).first

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
return

}else{
	utils.SendError(c,http.StatusUnauthorized,"verifying email code is incorrect")
	return
}
}

 
func Login(c *gin.Context){
	var input login
	utils.ParseANDSendResponse(c,&input)
if !utils.CheckEMail(input.Email){
	utils.SendError(c,http.StatusForbidden,"enter valid email")
	return
}

if utils.Attempts(c,input.Email){
return
}
var user model.Users


if res:=config.DB.Where("email = ?",input.Email).First(&user);!utils.CheckPass(input.Password,user.PasswordHash)|| res.RowsAffected==0||res.Error!=nil {
	 if !utils.IncrAttempts(c,input.Email){
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
}else {
	utils.SendRes(c,"email is verifyed u good to go")
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
	if subtle.ConstantTimeCompare([]byte(code),[]byte(input.Code))==1{
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
		var users model.Users
	num:= config.DB.Model(&model.Users{}).Where("email=? AND (tfa_verifed=? OR Login_codes_set=?)",input.Email,true,true).Find(&users).RowsAffected
	if num!=0{
utils.SendRes(c,"Enter your 2FA code to login OR login codes to enter")
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
utils.SetDeviceInfo(c,input.Email)
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
error:=config.DB.Model(&model.Users{}).Where("email=?",input.Email).Update("tfa_code",secret).Error
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
utils.SetDeviceInfo(c,input.Email)
utils.SendRes(c,response)

    

	}
}
func GenerteLoginCodes( c *gin.Context){
var	input Email
	var codes []string
	if !utils.ParseANDSendResponse(c,&input){
		return
	}
	fmt.Println(input.Email)
	for i:=0;i<12;i++{
		n,err:=rand.Int(rand.Reader,big.NewInt(1000000))
		if err!=nil{
			utils.SendError(c,http.StatusInternalServerError,"something went wrong")
			return
		}
		codes = append(codes,fmt.Sprintf("%06s",n.String()) )


	} 
	err:=config.DB.Model(&model.Users{}).Where("email=?",input.Email).Update("login_codes",strings.Join(codes, ",")).RowsAffected
			if err==0{
			utils.SendError(c,http.StatusInternalServerError,"something went wrong")
			errs:=config.DB.Model(&model.Users{}).Where("email=?",input.Email).RowsAffected
			fmt.Println(errs)
			return
			}
Messga:=fmt.Sprintf(`Hello [UserName],

🛡️ You’ve successfully generated backup login codes for your account security.

These codes can be used in place of 2FA in case you lose access to your authenticator app or device.

🔢 Your Backup Login Codes:
--------------------------------
%s
%s
%s
%s
%s
%s
%s
%s
%s
%s
%s
%s
--------------------------------

Each code is valid **only once**, and should be stored in a safe place (like a password manager).

🚨 Keep in mind:
- Do NOT share these codes with anyone.
- If you think someone else has access to them, generate new ones immediately.
- Using a code will invalidate it — it cannot be reused.

Stay safe and secure,  
🔐 The SOAH Security Team
`,codes[0],codes[1],codes[2],codes[3],codes[4],codes[5],codes[6],codes[7],codes[8],codes[9],codes[10],codes[11])
errd:=config.DB.Model(&model.Users{}).Where("email=?",input.Email).Update("login_codes_set",true).Error
if errd!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	fmt.Println(errd)
	return
}
utils.SendEmailSmtp(c,input.Email,Messga)

utils.SendRes(c,Messga)
}

func VerifyLoginCode(c *gin.Context){
	var input Verify
	if ! utils.ParseANDSendResponse(c,&input){return}
	var user model.Users
	err:=config.DB.Model(&model.Users{}).Where("email=?",input.Email).First(&user).Error
	if err!=nil{
		utils.SendError(c,http.StatusInternalServerError,"something went error")
		return
	}
	//codes:=strings.Split(user.Login_codes,",")
	codes:=strings.Split(user.Login_codes, ",")
	var found bool
	var copy []string
	for i:=0;i<len(codes);i++{
if subtle.ConstantTimeCompare([]byte(codes[i]),[]byte(input.Code))==1{
found =true
continue
}
copy = append(copy, codes[i])
	}
	if found{
	// bytes,err3:=config.Rdb.Get(config.Ctx,"Login:user:"+input.Email).Result()
	// if err3!=nil{
	// 	utils.SendError(c,http.StatusInternalServerError,"sfsdsomething went wrong")
	// 	fmt.Println("wee",err3.Error())
	// 	return
	// }
// var user model.Users
// errr:=json.Unmarshal([]byte(bytes),&user)
// 	if errr!=nil{
// 		utils.SendError(c,http.StatusInternalServerError,"dassomething went wrong")
// 		return
// 	}
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
utils.SetDeviceInfo(c,input.Email)
utils.SendRes(c,response)

user.Login_codes=strings.Join(copy, ",")
err:=config.DB.Model(&model.Users{}).Where("email=?",input.Email).Update("Login_codes",user.Login_codes).Error
if err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	return
}
	
}else{
		utils.SendError(c,http.StatusUnauthorized,"Enter your Login_codes correctly")
		return
	}

}

func ForgetPassword(c *gin.Context){
var input Email 
if ! utils.ParseANDSendResponse(c,&input){
	return
}
token:=uuid.New().String()
email:=fmt.Sprintf("%s%s&email=%s", os.Getenv("Gmail_Forget_password"),token,input.Email)
message:=fmt.Sprintf(`Hi %s,

We received a request to reset the password for your account linked to this email: %s.

To proceed, click the secure link below:

🔗 Reset Link:
%s

This link will expire in 15 minutes for your protection.

If you did not request a password reset, please ignore this message or contact our support team.

Stay secure,  
The SOAH Team 🛡️
`,strings.TrimSuffix(input.Email,"@gmail.com"),input.Email,email)
config.Rdb.Set(config.Ctx,"Reset:token:"+input.Email,token,15*time.Minute)
utils.SendEmailSmtp(c,input.Email,message)
utils.SendRes(c,"Token sent by email check your email")
}
func ResetPassword(c *gin.Context){
	var input Reset
	if ! utils.ParseANDSendResponse(c,&input){
		return
	}
	token,err:=config.Rdb.Get(config.Ctx,"Reset:token:"+input.Email).Result()
	if err!=nil||err==redis.Nil{
		utils.SendError(c,http.StatusInternalServerError,"something went wrong")
		return
	}
	if subtle.ConstantTimeCompare([]byte(token),[]byte(input.Token))==1{
		hashed,err:=bcrypt.GenerateFromPassword([]byte(input.Password),bcrypt.DefaultCost)
		if err!=nil{
			utils.SendError(c,http.StatusInternalServerError,"something went wrong")
			return
		}
	if err:=config.DB.Model(&model.Users{}).Where("email=?",input.Email).Update("password_hash",hashed).Error;err!=nil{
		utils.SendError(c,http.StatusInternalServerError,"something went wrong")
		return
	}
	if err:=config.Rdb.Del(config.Ctx,"Reset:token:"+input.Email).Err();err!=nil{
		utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	     return	
	}
	utils.SendRes(c,"password Reseted correctly")
	}else{
		utils.SendError(c,http.StatusUnauthorized,"unathourized access cannot reset password")
		return
	}
}

func GetDeviceInfo(c *gin.Context){
	email:=c.Query("email")
var user model.Users
err:=config.DB.Model(&model.Users{}).Where("email=?",email).Find(&user).Error
if err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	fmt.Println("623",err.Error())

	return
}
fmt.Println("user:",user)
var res model.DeviceRecord
err2:=config.DB.Model(&model.DeviceRecord{}).Where("userID=?",user.ID).Find(&res).Error
if err2!=nil{
	utils.SendError(c,http.StatusInternalServerError,"somethign went wrong")
	fmt.Println("AZA",err2.Error())
	return
}
fmt.Println(res)
utils.SendRes(c,res)

}
func ReauthPassword(c *gin.Context){
	var input login
if !utils.ParseANDSendResponse(c,&input){
		return
	}
	if !utils.CheckEMail(input.Email){
		utils.SendError(c,http.StatusUnauthorized,"invaild Email")
		return
	}
	var user model.Users
	var found bool
	if err:=config.DB.Model(&model.Users{}).Where("email=?",input.Email).First(&user).RowsAffected;err==0{
	found =false
		
	}else 
	{
		found =true
	}
if !found||!utils.CheckPass(input.Password,user.PasswordHash){
utils.SendError(c,http.StatusUnauthorized,"Enter Email or password correctly")
return
	}

if err:=config.Rdb.Set(config.Ctx,"Reauth:"+input.Email,"1",5*time.Minute).Err();err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
}
	c.Set("email",input.Email)
	utils.SendRes(c,"you can change your cerditional now")

}
func ReauthTFA(c *gin.Context){

	var input Verify
	if !utils.ParseANDSendResponse(c,&input){
		return
	}
	var user model.Users
	
	 res:=config.DB.Model(&model.Users{}).Where("email=?",input.Email).First(&user)
	 if res.Error!=nil||res.RowsAffected==0 {
		utils.SendError(c,http.StatusUnauthorized,"email isnot in system")
		return
	}
if !user.Tfa_verifed{

	utils.SendError(c,http.StatusForbidden,"you cannot use 2FA method it must be enables first")
	return
}
	code:=user.Tfa_code
	

	if!totp.Validate(input.Code,code){
		utils.SendError(c,http.StatusUnauthorized,"code isnot correct try again")
		return
	}
	if err:=config.Rdb.Set(config.Ctx,"Reauth:"+input.Email,"1",5*time.Minute).Err();err!=nil{
		utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	}
		c.Set("email",input.Email)
	utils.SendRes(c,"you can change your cerditional now")

}
func ReauthCode(c *gin.Context){

	var input Verify
		if utils.ParseANDSendResponse(c,&input){
		return
	}
	var user model.Users
	if err:=config.DB.Model(&model.Users{}).Where("email=?",input.Email).First(&user).RowsAffected;err==0{
		utils.SendError(c,http.StatusUnauthorized,"email isnot in system")
		return
	}
	if !user.Login_codes_set{
		utils.SendError(c,http.StatusForbidden,"you cannot use this codes method it must be enables first")
		return
	}
	var copy []string
	var found bool
	codes:=strings.Split(user.Login_codes, ",")
	for i:=0;i<len(codes);i++{
	if codes[i]==input.Code{
		found =true
		continue
	}
	copy = append(copy, codes[i])
	}
	if !found{
		utils.SendError(c,http.StatusUnauthorized,"Enter code correctly try again")
		return
	}
	if err:=config.DB.Model(&model.Users{}).Where("email=?",input.Email).Update("login_codes",strings.Join(copy, ",")).Error;err!=nil{
		utils.SendError(c,http.StatusInternalServerError,"something went wrong")
		return
	}
	
	
if err:=config.Rdb.Set(config.Ctx,"Reauth:"+input.Email,"1",5*time.Minute).Err();err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
}
	c.Set("email",input.Email)
utils.SendRes(c,"you can change your cerditional now")


}

func ChangePassword(c *gin.Context){
var input Password
if !utils.ParseANDSendResponse(c,input){
	return
}
if res:=utils.ValidatePassword(input.Password);res!="0"{
	utils.SendError(c,http.StatusBadRequest,res)
	return
}
email,exist:=c.Get("email")
if !exist{
	utils.SendError(c,http.StatusUnauthorized,"you are unauthorized to enter this route")
	return
}
id,exist:=c.Get("id")
if !exist{
	utils.SendError(c,http.StatusUnauthorized,"you are unauthorized to enter this route")
	return
}
if res:=utils.NotOldPassword(input.Password,id.(uint));res!="0"{
	utils.SendError(c,http.StatusBadRequest,res)
}
if subtle.ConstantTimeCompare([]byte(input.Password),[]byte(input.Confirm))!=1{
	utils.SendError(c,http.StatusUnauthorized,"confirm password isnot like the password")
	return
}
hashed:=utils.HashPassword(c,input.Password)
if err:=config.DB.Model(&model.Users{}).Where("email=?",email.(string)).Update("password_hash",hashed).Error;err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	return
}
username,exist:=c.Get("Username")
if !exist{
	utils.SendError(c,http.StatusUnauthorized,"you are unauthorized to enter this route")
	return
}

data,err:=utils.Sendlocation(c.Request.RemoteAddr)
if err!=nil{
	utils.SendError(c,http.StatusInternalServerError,"something went wrong")
	return
}
message:=fmt.Sprintf(`Hi ,%s

We’re letting you know that the password for your account (**%s*) was just changed.

🕒 Time: %s
📍 Location: %s,%s 
🌐 IP Address: %s
🖥️ Device: %s

---

If **you made this change**, no further action is needed.  
If **you did NOT** change your password, please secure your account immediately:

🔒 [Reset your password now]  
🔍 [Review recent account activity]

Your security is our top priority.  
If you have any questions, reply to this email or contact our support team.

Stay safe,  
**The SOAH Team**
`,username,email.(string),data.Timezone,data.City,data.Country,data.Query,c.Request.UserAgent())
utils.SendEmailSmtp(c,email.(string),message)
utils.SendRes(c,"password updated correctly")

}

func ChangeEmail(c *gin.Context){
	var input Email
	
	if !utils.ParseANDSendResponse(c,input){
		return
	}
if !utils.CheckEMail(input.Email)||! utils.VerifEmailHelper(c,input.Email){
	utils.SendError(c,http.StatusUnauthorized,"input email is not valid")
}



	
}
func VerifyNewEmail(c *gin.Context){
  var	input Verify
    email,exist:=c.Get("email")
	if !exist{
		utils.SendError(c,http.StatusUnauthorized,"you cannot use this codes method it must be enables first")
		return
	}
	 code,err:=config.Rdb.Get(config.Ctx,"ChangeEmail:code:"+email.(string)).Result()
	 if err!=nil{
		utils.SendError(c,http.StatusInternalServerError,"something went wrong")
		return
	 }
	 if subtle.ConstantTimeCompare([]byte(code),[]byte(input.Code))!=1{
		utils.SendError(c,http.StatusUnauthorized,"Enter code correctly")
		return
	 }
 if err:=config.DB.Model(&model.Users{}).Where("email=?",email).Update("email",input.Email).Error;err!=nil{
		utils.SendError(c,http.StatusInternalServerError,"something went wrong")
		return
	 }
   utils.SendRes(c,"email changed correctly")
}
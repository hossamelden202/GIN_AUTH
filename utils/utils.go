package utils

import (
	"GIN/config"

	"GIN/model"

	"fmt"

	//"fmt"
	//"GIN/utils"
"net/smtp"
	"os"
	"regexp"
	"strings"
	"time"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func GenerteJwt(Username string,Email string ,id int,role string,Time time.Duration)(string,error){
token:=jwt.NewWithClaims(jwt.SigningMethodHS256,jwt.MapClaims{
	"Username":Username,
	"email":Email,
	"role":role,
	"id":id,
	"exp":time.Now().Add(Time).Unix(),
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
func CheckPass(password string,hashed string)bool {
	err:=bcrypt.CompareHashAndPassword([]byte(hashed),[]byte(password))
	if err!=nil{
		return false
	}
	return true
}
func VaildateToken(c *gin.Context,tokenstring string)(*jwt.Token,bool){
	secret:=[]byte(os.Getenv("jwt_secret"))
token,err:= jwt.Parse(tokenstring,func(token *jwt.Token)(interface{},error){

if _,ok:=token.Method.(*jwt.SigningMethodHMAC);!ok{
	return nil,jwt.ErrSignatureInvalid

}
return secret,nil
})

if err!=nil||!token.Valid{
SendError(c,http.StatusUnauthorized,"invalid token")
return nil,false
}
if claims,ok:=token.Claims.(jwt.MapClaims);ok{
	if val,ok:=claims["id"].(float64);ok{
	c.Set("id",int(val))
	}
	if val,ok:=claims["username"];ok{
		c.Set("username",val)
	}
	if val,ok:=claims["email"];ok{
		c.Set("email",val)
	}
	if val,ok:=claims["role"];ok{
		c.Set("role",val)
	}
}else {
	return nil,false
}
return token,true
}

func Attempts(c *gin.Context,email string)bool{
	fmt.Println("just wondering")
str2,err2:=config.Rdb.Get(config.Ctx,"Login:block:"+email).Result()
	if err2==nil&&err2==redis.Nil{
	SendError(c,http.StatusInternalServerError,"Email isnot blocked")
		return false
	}
	if str2=="1"{
		SendError(c,http.StatusTooManyRequests,"you are blocked wait 15 min utils you can try again")
		return true
	}

	if num,err:=config.Rdb.Exists(config.Ctx,"Login:fail:"+email).Result();err!=nil||num==0{
		fmt.Println("First try to login")
		return false
	}
num,err:=config.Rdb.Get(config.Ctx,"Login:fail:"+email).Result()
if err!=nil{
	
	SendError(c,http.StatusInternalServerError,"something just went2 wrong")
	return true
}
num2,err2:=strconv.Atoi(num)
if err2!=nil{
	SendError(c,http.StatusInternalServerError,"something just went3 wrong")
	return true
}

ttl,errT:=config.Rdb.TTL(config.Ctx,"Login:fail:"+email).Result()
if errT!=nil{
	SendError(c,http.StatusInternalServerError,"something went wrong")
	return true
}////captcha-solved?email=hossam2@gmail.com
if num2==3{
	num2++
	err:=config.Rdb.Set(config.Ctx,"Login:fail:"+email,num2,ttl).Err()
	fmt.Println(ttl ,"CC is here ",num2)
	if err!=nil{
		SendError(c,http.StatusInternalServerError,"something went wrong")
		return true
	}
	SendError(c,http.StatusUnauthorized,"Slove Captcha first")
	
	return true
}
if num2>3&&num2<5{
	num,err:=config.Rdb.Exists(config.Ctx,"captcha:passed:"+email).Result()

	if err!=nil||num==0{
SendError(c,http.StatusUnauthorized,"You should Solve Captcha First")
return true
	}
	return false
}
 fmt.Println(num2 ,"GG here is ")
if num2>=5{
	fmt.Println("in checking if u can login found out num of attempts is",num2,ttl)

    SendEmail(c,email)

	config.Rdb.Set(config.Ctx,"Login:block:"+email,"1",5*time.Second)
	SendError(c,http.StatusUnauthorized,fmt.Sprintf("You exceeded number of Attempts wait for %v",ttl))
	
	return true
} 
return false
}
func IncrAttempts(c *gin.Context,email string)bool{
	// var err error
	// var num int64
	num,err:=config.Rdb.Exists(config.Ctx,"Login:fail:"+email).Result();
	if err!=nil||num==0{
		//fmt.Println("here uis eror",err,num);
	fmt.Println("doesnt exist so add it")
	err:=config.Rdb.Set(config.Ctx,"Login:fail:"+email,0,time.Second*5).Err()
	if err!=nil{
	SendError(c,http.StatusInternalServerError,"something went wrong 6")
	return false
	}
}

	str,err:=config.Rdb.Get(config.Ctx,"Login:fail:"+email).Result()
	//fmt.Println(str,"here is the error",err.Error())
	if err!=nil{
		SendError(c,http.StatusInternalServerError,"something just went4 wrong")

		return false
	}
	str=strings.TrimSpace(str)
	nume,err2:=strconv.Atoi(str)
	if err2!=nil{
		SendError(c,http.StatusInternalServerError,"something just went5 wrong")
		return false
	}
	nume++
	fmt.Println(nume)
	ttl,errT:=config.Rdb.TTL(config.Ctx,"Login:fail:"+email).Result()
	if errT!=nil{
   SendError(c,http.StatusInternalServerError,"something went wrong")
return false
	}
	if err:=config.Rdb.Set(config.Ctx,"Login:fail:"+email,nume,ttl).Err();err!=nil{
		SendError(c,http.StatusInternalServerError,"something went wrong")
		return false
	}
	 return true

}
func SendEmail(c *gin.Context,email string){
Addr:=c.Request.RemoteAddr
fmt.Println("hello gere is addr",Addr)
var user model.Users
//email:="kittc584@gmail.com"
config.DB.Where("email=?",email).First(&user)
Message:=fmt.Sprintf(
`Hello %s,

We noticed multiple failed login attempts on your account using the email: %s.

📍 Location (based on IP): %s
🕒 Time: %v

As a security precaution, we’ve temporarily blocked login from this IP for 15 minutes.

If this wasn’t you, we recommend:
- Changing your password immediately.
- Enabling extra security options, like 2FA .

If you recognize this activity, no action is required.

Stay safe,
We onto you Nigga
The Racist Team 🛡️
`,user.Username,user.Email,Addr,time.Now())

SendEmailSmtp(c,email,Message)

}
func SendEmailSmtp(c *gin.Context,email string,Message string ){
	port:="587"

	host:="smtp.gmail.com"
	auth:=smtp.PlainAuth("",os.Getenv("Mail_email"),os.Getenv("Mail_password"),host)
	err:=smtp.SendMail(host+":"+port,auth,os.Getenv("Mail_email"),[]string{email},[]byte (Message))
	if err!=nil{
		SendError(c,http.StatusInternalServerError,fmt.Sprintf("something went wrong:%s",err))
		return
	}
}



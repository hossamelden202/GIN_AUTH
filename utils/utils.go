package utils

import (
	"GIN/config"
	"encoding/json"
	// "html"
	"io"
	"net"

	"GIN/model"

	"fmt"

	//"fmt"
	//"GIN/utils"
	//"github.com/mailersend/mailersend-go"

	"os"
	"regexp"
	"strings"
	"time"

	"net/http"
	"net/smtp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)
type GeoData struct {
	Query        string `json:"query"`
	Country      string `json:"country"`
	RegionName   string `json:"regionName"`
	City         string `json:"city"`
	ISP          string `json:"isp"`
	Timezone     string `json:"timezone"`
	Org          string `json:"org"`
	Status       string `json:"status"`
}

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

   SendEmail_FAILED_LOGIN(c,email)

	config.Rdb.Set(config.Ctx,"Login:block:"+email,"1",5*time.Second)
	SendError(c,http.StatusUnauthorized,fmt.Sprintf("You exceeded number of Attempts wait for %v",ttl))
	
	return true
} 
return false
}
func RestAttempts(email string){
	if num,err:=config.Rdb.Exists(config.Ctx,"Login:fail:"+email).Result();err!=nil||num!=0{
config.Rdb.Del(config.Ctx,"Login:fail:"+email)
	}
		if num,err:=config.Rdb.Exists(config.Ctx,"Login:block:"+email).Result();err!=nil||num!=0{
config.Rdb.Del(config.Ctx,"Login:block:"+email)
	}
	
	
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
func SendEmail_FAILED_LOGIN(c *gin.Context,email string){
Addr:=c.Request.RemoteAddr
geo,err:=Sendlocation(Addr)
if err!=nil{SendError(c,http.StatusInternalServerError,"something went wrong")
fmt.Println(err.Error())
return}
fmt.Println("hello gere is addr",Addr)
 var user model.Users
// email:="kc334844@gmail.com"
config.DB.Where("email=?",email).First(&user)
Message := fmt.Sprintf(
`Hello %s,

We noticed multiple failed login attempts on your account using the email: %s.

📍 Location Details:
- IP Address: %s
- Country: %s
- Region: %s
- City: %s
- ISP: %s
- Organization: %s 
- Timezone: %s
🕒 Time: %v

As a security precaution, we’ve temporarily blocked login from this IP for 15 minutes.

If this wasn’t you, we recommend:
- Changing your password immediately.
- Enabling extra security options, like 2FA.

If you recognize this activity, no action is required.

Stay safe,  
The Racist Team 🛡️
`,
	user.Username,         //  — username
	user.Email,            // %s — email
	geo.Query,             // %s — IP
	geo.Country,           // %s — country
	geo.RegionName,        // %s — region
	geo.City,              // %s — city
	geo.ISP,               // %s — ISP
	geo.Org,               // %s — organization
	geo.Timezone,          // %s — timezone
	time.Now(),            // %v — timestamp
)

SendEmailSmtp(c,email,Message)

}
func SendEmail(c *gin.Context,email string){
	Addr:=c.Request.RemoteAddr
geo,err:=Sendlocation(Addr)
if err!=nil{SendError(c,http.StatusInternalServerError,"something went wrong")
fmt.Println(err.Error())
return}
fmt.Println("hello gere is addr",Addr)
 var user model.Users
// email:="kc334844@gmail.com"
config.DB.Where("email=?",email).First(&user)
Message := fmt.Sprintf(
`Subject: 🎉 Welcome to Racist Team, %s!

Hello %s 👋,

Welcome aboard! Your account with the email: %s has just been created successfully.

🗺️ Location at Signup:
- IP Address: %s
- Country:%s
- Region: %s
- City: %s
- ISP: %s
- Organization: %s
- Timezone: %s

🕒 Signup Time: %v

We’re thrilled to have you in the Racist Team family.  
Feel free to explore, connect, and enjoy everything we’ve built for you 💥

If you didn’t create this account, please contact us immediately.

Cheers,  
The Racist Team 🛡️

`,
	user.Username,
	user.Username,         //  — username
	user.Email,            // %s — email
	geo.Query,             // %s — IP
	geo.Country,           // %s — country
	geo.RegionName,        // %s — region
	geo.City,              // %s — city
	geo.ISP,               // %s — ISP
	geo.Org,               // %s — organization
	geo.Timezone,          // %s — timezone
	time.Now(),            // %v — timestamp
)

SendEmailSmtp(c,email,Message)
}
func SendEmailSmtp(c *gin.Context,email string,Message string ){
	
// 	ms:=mailersend.NewMailersend(os.Getenv("Ms_API_KEY"))
// 	msg:=ms.Email.NewMessage()
// 	From:=mailersend.From{Name:"team",Email: os.Getenv("Mail_email")}
// 	to:=mailersend.Recipient{Name:userName,Email: email}
// 	msg.SetFrom(From)
// 	msg.SetRecipients([]mailersend.Recipient{to})
// 	msg.SetSubject("Alert: Login issue")
// 	// msg.SetText(Message)
// 	msg.SetHTML("<pre>"+html.EscapeString(Message)+"</pre>")
// fmt.Println("EMAIL MESSAGE:\n", Message)

// 	_,_,err:=ms.BulkEmail.Send(config.Ctx,[]*mailersend.Message{msg})

auth:=smtp.PlainAuth("",os.Getenv("Mail_email"),os.Getenv("Mail_password"),"smtp.gmail.com")
    Addr:="smtp.gmail.com"+":"+"587"
	err:=smtp.SendMail(Addr,auth,os.Getenv("Mail_email"),[]string{email},[]byte(Message))
	

	if err!=nil{
		
		SendError(c,http.StatusInternalServerError,fmt.Sprintf("something went wrong:%s",err))
		return
	}
}
func Sendlocation(ip string)(*GeoData,error){
	// fmt.Println("first",ip)
	
ip,_,err2:=net.SplitHostPort(ip)
if err2!=nil{return nil,err2}
// fmt.Println("sec",ip)
resp,err:=http.Get("http://ip-api.com/json/"+"8.8.8.8")



if err!=nil{return nil,err}
body,error:=io.ReadAll(resp.Body)
defer resp.Body.Close()
if error!=nil{return nil,error}
var data *GeoData
fmt.Println(resp)
fmt.Println("body:",string(body))
//can use parseAndsend written funcion  
error=json.Unmarshal(body,&data)
if error!=nil{return nil,error}
if data.Status!="success"{return nil,fmt.Errorf("failed to get geo info")}
return data,nil
}
//745058958154756


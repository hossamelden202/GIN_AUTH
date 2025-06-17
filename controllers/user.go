package controllers

import (
	//"net/http"

	"github.com/gin-gonic/gin"
	// "github.com/gin-gonic/gin/binding"
	"GIN/config"
	"GIN/model"
	"GIN/utils"
)


type input struct{
 Name string `json:"name" binding:"required,min=2,max=255,alpha_space"`
 Email string `json:"email" binding:"required,email"`
Country string `json:"Country" binding:"required"`
ZipCode string `json:"ZipCode" binding:"required"`
Phone string `json:"phone" binding:"required,numeric"`
Password string `json:"password" binding:"required"`
Gender string `json:"gender" binding:"required,oneof=male female"`

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
utils.SendError(c,errD)
return}
	Token,errT:=utils.GenerteJwt(username,input.Email,int(users.ID),users.Role)
	if errT!=nil{
	utils.SendError(c,errT)
return}
	

res:=struct{
User model.Users `json:"user"`
Token string `json:"token"`
}{
User:users,
Token:Token,
}
utils.SendRes(c,res)
}
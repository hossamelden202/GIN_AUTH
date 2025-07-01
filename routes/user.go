package routes

import (
	"GIN/controllers"
	"GIN/middlerware"
	"GIN/utils"

	"github.com/gin-gonic/gin"
)
func Routing(r *gin.Engine){

userRoutes:=r.Group("/user")
{

	userRoutes.POST("/signup",controllers.Create)
 userRoutes.POST("/login",controllers.Login)	

}
userRoutes.POST("/captcha-solved",controllers.Captcha)
userRoutes.POST("/refresh",controllers.Refresh)

userRoutes.Use(middlerware.Auth())
userRoutes=r.Group("/user")
{
userRoutes.POST("/logout",controllers.Logout)
userRoutes.POST("/test",utils.SendEmail)

}


}

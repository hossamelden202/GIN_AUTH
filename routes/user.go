package routes

import (
	"GIN/controllers"
	"GIN/middlerware"


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
userRoutes.POST("/verify-email",controllers.Verify_email)
userRoutes.POST("/create-2fA",controllers.Create_2fA)
userRoutes.POST("verify-2fA",controllers.Verify_2fA)
userRoutes.Use(middlerware.Auth())
userRoutes=r.Group("/user")
{
userRoutes.POST("/logout",controllers.Logout)


}


}

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
userRoutes.POST("/verify-2fA",controllers.Verify_2fA)
userRoutes.POST("/verify_signup_email",controllers.Verify_signup_email)
userRoutes.POST("/generte_login_codes",controllers.GenerteLoginCodes)
userRoutes.POST("/verify_login_code",controllers.VerifyLoginCode)
userRoutes.POST("/forget_password",controllers.ForgetPassword)
userRoutes.POST("/reset_password",controllers.ResetPassword)
userRoutes.GET("/device-info",controllers.GetDeviceInfo)
userRoutes.POST("/reauth-password",controllers.ReauthPassword)
userRoutes.POST("/reauth-2fa",controllers.ReauthTFA)
userRoutes.POST("/reauth-code",controllers.ReauthCode)
userRoutes.POST("/change-password",controllers.ChangePassword)
userRoutes.POST("/change-email",controllers.ChangeEmail)
userRoutes.POST("/verify-new-email",controllers.VerifyNewEmail)
userRoutes.Use(middlerware.Auth())

userRoutes=r.Group("/user")
{
userRoutes.POST("/logout",controllers.Logout)


}


}

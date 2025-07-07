package routes

import (
	"GIN/controllers"
	"GIN/middlerware"


	"github.com/gin-gonic/gin"
)
func Routing(r *gin.Engine){

userRoutes:=r.Group("/user")
{

userRoutes.POST("/signup",middlerware.GeoGurd(),controllers.Create)
userRoutes.POST("/login",middlerware.GeoGurd(),controllers.Login,middlerware.ActivateUser())	

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
userRoutes.POST("/reauth-password",middlerware.Auth(),controllers.ReauthPassword)
userRoutes.POST("/reauth-2fa",middlerware.Auth(),controllers.ReauthTFA)
userRoutes.POST("/reauth-code",middlerware.Auth(),controllers.ReauthCode)
userRoutes.POST("/change-password",middlerware.Auth(),middlerware.Reauth(),controllers.ChangePassword,middlerware.AfterChange(),middlerware.DeactivateUser())
userRoutes.POST("/change-email",middlerware.Auth(),middlerware.Reauth(),controllers.ChangeEmail,middlerware.DeactivateUser())
userRoutes.POST("/verify-new-email",middlerware.Auth(),controllers.VerifyNewEmail)

userRoutes.GET("/me",middlerware.Auth(),controllers.GetUser)
userRoutes.GET("/get-sessions",middlerware.Auth(),controllers.GetSession)
userRoutes.POST("/logout-all",middlerware.Auth(),controllers.LogoutALL)
userRoutes.POST("/logout/:sessionid",middlerware.Auth(),controllers.LogoutSession)
userRoutes=r.Group("/user")
userRoutes.Use(middlerware.Auth())
{
userRoutes.POST("/logout",controllers.Logout,middlerware.DeactivateUser())


}


}

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


}
userRoutes.Use(middlerware.Auth())
userRoutes.POST("/login",controllers.Login)	

}

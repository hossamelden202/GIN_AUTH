package routes

import (
	"GIN/controllers"

	"github.com/gin-gonic/gin"
)
func Routing(r *gin.Engine){

userRoutes:=r.Group("/user")
{

	userRoutes.POST("/signup",controllers.Create)


}
}

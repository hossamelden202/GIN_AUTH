package main

import (
	"GIN/config"
	"GIN/routes"
	//"GIN/utils"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// During your main setup or initialization function:
func registerCustomValidators() {
  if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
    v.RegisterValidation("alpha_space", func(fl validator.FieldLevel) bool {
      value := fl.Field().String()
      regex := regexp.MustCompile(`^[a-zA-Z\s]+$`) // letters + spaces only
      return regex.MatchString(value)
    })
  }
}

func main(){
config.Connect()
config.ConnectRedis()
r:=gin.Default()
registerCustomValidators()
routes.Routing(r)

r.Run(":8080")
}
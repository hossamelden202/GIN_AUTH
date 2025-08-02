package main

import (
	"GIN/config"
	"GIN/routes"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// registerCustomValidators registers custom validation rules
func registerCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("alpha_space", func(fl validator.FieldLevel) bool {
			value := fl.Field().String()
			regex := regexp.MustCompile(`^[a-zA-Z\s]+$`) // letters + spaces only
			return regex.MatchString(value)
		})
	}
}

// serveSwaggerUI serves the Swagger UI documentation
func serveSwaggerUI(c *gin.Context) {
	c.File("./docs/index.html")
}

// serveSwaggerYAML serves the Swagger YAML specification
func serveSwaggerYAML(c *gin.Context) {
	c.File("./docs/swagger.yaml")
}

func main() {
	// Initialize database and cache
	config.Connect()
	config.ConnectRedis()

	// Create Gin engine
	r := gin.Default()

	// Register custom validators
	registerCustomValidators()

	// Serve static documentation files
	r.GET("/api-docs/", serveSwaggerUI)
	r.GET("/docs/swagger.yaml", serveSwaggerYAML)

	// Serve Swagger UI assets from CDN (no need to host locally)
	r.StaticFile("/favicon.ico", "./docs/favicon.ico")

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "GIN Auth Microservice",
		})
	})

	// Setup all API routes
	routes.Routing(r)

	// Start server
	r.Run(":8080")
}

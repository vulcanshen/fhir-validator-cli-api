package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	_ "github.com/vulcanshen/fhir-validator-cli-api/docs"
	"github.com/vulcanshen/fhir-validator-cli-api/handler"
	"log"
)

var port = 8081

// @title FHIR Validator API
// @version 1
// @Description http api for fhir validator_cli
// @contact.name Vulcan Shen
// @contact.url https://github.com/vulcanshen
// @contact.email vulcan.shen.2304@gmail.com
// @host localhost:8081
// @BasePath /api
func main() {
	router := gin.Default()

	api := router.Group("/api")

	validation := api.Group("/validate")
	{
		validation.POST("", handler.SyncValidation)
		validation.POST("/sse", handler.SseValidation)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Fatal(router.Run(fmt.Sprintf(":%d", port)))
}

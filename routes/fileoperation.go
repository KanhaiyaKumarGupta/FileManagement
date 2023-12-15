package routes

import (
	"github.com/KanhaiyaKumarGupta/jwt-authentication/controllers"
	"github.com/KanhaiyaKumarGupta/jwt-authentication/middleware"
	"github.com/gin-gonic/gin"
)

func FileRoutes(router *gin.Engine) {
	router.Use(middleware.Authenticate())

	router.POST("/uploadfiles", controllers.UploadFile())
	router.GET("/downloadfiles", controllers.DownloadFile())
	router.GET("/fetchtransactions", controllers.FetchTransactions())
	router.DELETE("/deletefiles", controllers.DeleteFile())
	router.GET("/listallfiles", controllers.ListFiles())
}

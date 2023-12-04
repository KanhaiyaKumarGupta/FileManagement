package routes

import (
	"github.com/KanhaiyaKumarGupta/jwt-authentication/controllers"
	"github.com/gin-gonic/gin"
)

func Authrouter(routes *gin.Engine) {
	routes.POST("/users/signup", controllers.Signup())
	routes.POST("/users/login", controllers.Login())
}

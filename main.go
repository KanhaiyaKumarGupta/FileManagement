package main

import (
	"log"
	"os"

	"github.com/KanhaiyaKumarGupta/jwt-authentication/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := createUploadsDir("./uploads"); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())

	routes.Authrouter(router)
	routes.FileRoutes(router)
	router.GET("api-1", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "Access granted for api 1"})
	})
	router.GET("api-2", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "Access granted for ap1-2"})
	})
	router.Run(":" + port)
}

func createUploadsDir(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.Mkdir(dirPath, 0755)
	}
	return nil
}

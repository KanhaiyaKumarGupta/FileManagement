package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/KanhaiyaKumarGupta/jwt-authentication/databases"
	"github.com/KanhaiyaKumarGupta/jwt-authentication/helpers"
	"github.com/KanhaiyaKumarGupta/jwt-authentication/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = databases.OpenCollection(databases.Client, "user")
var validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("email of password is incorrect")
		check = false
	}
	return check, msg
}
func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		defer cancel()
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		count1, err := userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the phone number"})
		}

		if count > 0 || count1 > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone number already exists"})
			return
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		token, refreshToken, _ := helpers.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *&user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInsertionNumber)
	}

}
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		var founduser models.User

		err := c.BindJSON(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&founduser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or Passwprd is incorrect"})
			return
		}
		passWordIsInvalid, msg := VerifyPassword(*user.Password, *founduser.Password)
		if !passWordIsInvalid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		if founduser.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		}
		signedtoken, refreshToen, err := helpers.GenerateAllTokens(*founduser.Email, *founduser.First_name, *founduser.Last_name, *&founduser.User_id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		helpers.UpdateAllTokens(signedtoken, refreshToen, founduser.User_id)
		err = userCollection.FindOne(ctx, bson.M{"user_id": founduser.User_id}).Decode(&founduser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, founduser)

	}
}

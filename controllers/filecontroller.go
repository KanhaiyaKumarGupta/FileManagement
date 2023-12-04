package controllers

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/KanhaiyaKumarGupta/jwt-authentication/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func UploadFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileDir := "./uploads"
		if err := os.MkdirAll(fileDir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
			return
		}
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
			return
		}
		filepath := filepath.Join(fileDir, file.Filename)
		if err := c.SaveUploadedFile(file, filepath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save the file"})
			return
		}
		transaction := models.FileTransaction{
			FileName:   file.Filename,
			Operation:  "Upload",
			Size:       file.Size,
			AccessedAt: time.Now(),
		}
		_, err = userCollection.InsertOne(context.Background(), transaction)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to log the upload in MongoDB"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully!", "filename": file.Filename})
	}
}
func DownloadFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		var downloadReq struct {
			FileName string `json:"fileName"`
		}

		if err := c.BindJSON(&downloadReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		fileName := downloadReq.FileName
		if fileName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File name is required"})
			return
		}

		fileDir := "./uploads"
		filepath := filepath.Join(fileDir, fileName)

		fileInfo, err := os.Stat(filepath)
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		transaction := models.FileTransaction{
			FileName:   fileName,
			Operation:  "Download",
			Size:       fileInfo.Size(),
			AccessedAt: time.Now(),
		}

		_, err = userCollection.InsertOne(context.Background(), transaction)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to log the download in MongoDB"})
			return
		}

		c.File(filepath)
	}
}

func ListFiles() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileDir := "./uploads"

		files, err := ioutil.ReadDir(fileDir)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read directory"})
			return
		}

		sort.Slice(files, func(i, j int) bool {
			return files[i].Size() < files[j].Size()
		})

		var fileInfos []map[string]interface{}
		for _, file := range files {
			fileInfos = append(fileInfos, map[string]interface{}{
				"name": file.Name(),
				"size": file.Size(),
			})
		}
		transaction := models.FileTransaction{
			Operation:  "List",
			AccessedAt: time.Now(),
		}
		_, err = userCollection.InsertOne(context.Background(), transaction)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to log the list operation in MongoDB"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"files": fileInfos})
	}
}

func FetchTransactions() gin.HandlerFunc {
	return func(c *gin.Context) {
		var transactions []models.FileTransaction

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := userCollection.Find(ctx, bson.D{{}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch transactions"})
			return
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var transaction models.FileTransaction
			if err := cursor.Decode(&transaction); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding transaction data"})
				return
			}
			transactions = append(transactions, transaction)
		}

		if err := cursor.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cursor error encountered"})
			return
		}

		c.JSON(http.StatusOK, transactions)
	}
}
func DeleteFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		var deleteReq struct {
			FileName string `json:"fileName"`
		}

		if err := c.BindJSON(&deleteReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		fileName := deleteReq.FileName
		if fileName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File name is required"})
			return
		}

		fileDir := "./uploads"
		filepath := filepath.Join(fileDir, fileName)

		fileInfo, err := os.Stat(filepath)
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		err = os.Remove(filepath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete the file"})
			return
		}

		transaction := models.FileTransaction{
			FileName:   fileName,
			Operation:  "Delete",
			Size:       fileInfo.Size(),
			AccessedAt: time.Now(),
		}

		_, err = userCollection.InsertOne(context.Background(), transaction)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to log the delete operation in MongoDB"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully!"})
	}
}

package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func renewBookHandler(context *gin.Context) {
	iUserID, _ := context.Get("userID")
	userID := iUserID.(int)
	bookIDString := context.PostForm("bookID")
	bookID, _ := strconv.Atoi(bookIDString)
	borrowIDString := context.PostForm("borrowID")
	borrowID, _ := strconv.Atoi(borrowIDString)
	result := agent.RenewBook(borrowID, userID, bookID)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

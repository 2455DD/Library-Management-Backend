package main

import (
	"github.com/gin-gonic/gin"
	. "lms/services"
	"net/http"
)

func registerHandler(context *gin.Context) {
	username := context.PostForm("username")
	password := "1234"
	email := context.PostForm("email")
	user := User{
		Username: username,
		Password: password,
		Email:    email,
		Debt:     0,
	}
	registerResult := dbAgent.RegisterMember(&user)
	context.JSON(http.StatusOK, gin.H{"status": registerResult.Status, "msg": registerResult.Msg, "userId": user.UserId})
}

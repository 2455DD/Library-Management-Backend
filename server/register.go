package main

import (
	"github.com/gin-gonic/gin"
	. "lms/services"
	"net/http"
)

const defaultPassword = "1234"

func registerHandler(context *gin.Context) {
	username := context.PostForm("username")
	email := context.PostForm("email")
	user := User{
		Username: username,
		Password: defaultPassword,
		Email:    email,
		Debt:     0,
	}
	registerResult := dbAgent.RegisterMember(&user)
	context.JSON(http.StatusOK, gin.H{"status": registerResult.Status, "msg": registerResult.Msg, "userId": user.UserId})
}

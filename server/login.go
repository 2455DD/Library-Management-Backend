package main

import (
	"github.com/gin-gonic/gin"
	. "lms/services"
	"lms/util"
	"net/http"
	"strconv"
)

func loginHandler(context *gin.Context) {
	login(context, dbAgent.AuthenticateUser, util.UserKey)
}

func adminLoginHandler(context *gin.Context) {
	login(context, dbAgent.AuthenticateLibrarian, util.AdminKey)
}

func login(context *gin.Context, authFunc LoginInterface, key []byte) {
	userIdStr := context.PostForm("userId")
	password := context.PostForm("password")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		context.Status(http.StatusBadRequest)
	}
	loginResult := authFunc(userId, password)
	if loginResult.Status == LoginOK {
		token := util.GenToken(userId, key)
		context.JSON(http.StatusOK, gin.H{"status": loginResult.Status, "msg": loginResult.Msg, "token": token})
	} else {
		context.JSON(http.StatusOK, gin.H{"status": loginResult.Status, "msg": loginResult.Msg})
	}
}

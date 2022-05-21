package main

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func getBorrowBooksPagesHandler(context *gin.Context) {
	iUserID, _ := context.Get("userId")
	userID := iUserID.(int)
	context.JSON(http.StatusOK, gin.H{"page": dbAgent.GetMemberBorrowBooksPages(userID)})
}

func getBorrowBooksHandler(context *gin.Context) {
	iUserID, _ := context.Get("userId")
	userID := iUserID.(int)
	pageStr := context.PostForm("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	books := dbAgent.GetMemberBorrowBooks(userID, page)

	bf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(bf)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(books)

	_, _ = context.Writer.Write(bf.Bytes())
}

func getReserveBooksPagesHandler(context *gin.Context) {
	iUserID, _ := context.Get("userId")
	userID := iUserID.(int)
	context.JSON(http.StatusOK, gin.H{"page": dbAgent.GetMemberReserveBooksPages(userID)})
}

func getReserveBooksHandler(context *gin.Context) {
	iUserID, _ := context.Get("userId")
	userID := iUserID.(int)
	pageStr := context.PostForm("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	books := dbAgent.GetMemberReserveBooks(userID, page)

	bf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(bf)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(books)

	_, _ = context.Writer.Write(bf.Bytes())
}

func borrowBookHandler(context *gin.Context) {
	iUserID, _ := context.Get("userId")
	userID := iUserID.(int)
	bookIDString := context.PostForm("bookId")
	bookID, _ := strconv.Atoi(bookIDString)
	result := agent.BorrowBook(userID, bookID)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

func reserveBookHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	bookIdStr := context.PostForm("bookId")
	bookId, err := strconv.Atoi(bookIdStr)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	result := dbAgent.ReserveBook(userId, bookId)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

func cancelReserveBookHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	bookIdStr := context.PostForm("bookId")
	bookId, err := strconv.Atoi(bookIdStr)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	result := dbAgent.CancelReserveBook(userId, bookId)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

func getReaderDashboardHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	result := agent.CountHistoryBorrowedBooks(userId) //历史借阅的图书量
	borrowedBookCount := result.Msg
	result = agent.TotalFineAmount(userId) //总罚款数量
	Totalfine := result.Msg
	result = agent.CountHistoryReservedBooks(userId) //历史预定图书量
	reservedBookCount := result.Msg
	result = agent.LastReturnBook(userId) //最后归还的图书
	lastReturnBook := result.Msg
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "History Borrowed books": borrowedBookCount, "Total Fine Amount": Totalfine, "History Reserve books": reservedBookCount, "Last Return Books": lastReturnBook})
}

func getMemberHistoryBorrowTimeHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	days := dbAgent.GetMemberHistoryBorrowTime(userId)
	context.JSON(http.StatusOK, gin.H{"days": days})
}

func getMemberFineHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	fine := agent.GetMemberFine(userId)
	context.JSON(http.StatusOK, gin.H{"fine": fine})
}

func getMemberPayURLHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	url := agent.GetPayMemberFineURL(userId)
	context.JSON(http.StatusOK, gin.H{"url": url})
}

func updatePasswordHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	oldPassword, ok1 := context.GetPostForm("oldPassword")
	newPassword, ok2 := context.GetPostForm("newPassword")
	if !(ok1 && ok2) {
		context.Status(http.StatusBadRequest)
		return
	}
	result := agent.UpdatePassword(userId, oldPassword, newPassword)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

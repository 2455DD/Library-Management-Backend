package main

import (
	"github.com/gin-gonic/gin"
	. "lms/services"
	"lms/util"
	"net/http"
	"strconv"
)

func getBorrowBooksPagesHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	context.JSON(http.StatusOK, gin.H{"page": dbAgent.GetMemberBorrowBooksPages(userId)})
}

func getBorrowBooksHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	pageStr := context.PostForm("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	books := dbAgent.GetMemberBorrowBooks(userId, page)
	_, _ = context.Writer.Write(util.JsonEncode(books))
}

func getReserveBooksPagesHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	context.JSON(http.StatusOK, gin.H{"page": dbAgent.GetMemberReserveBooksPages(userId)})
}

func getReserveBooksHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	pageStr := context.PostForm("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	books := dbAgent.GetMemberReserveBooks(userId, page)
	_, _ = context.Writer.Write(util.JsonEncode(books))
}

func borrowBookHandler(context *gin.Context) {
	userId, err1 := strconv.Atoi(context.PostForm("userId"))
	bookId, err2 := strconv.Atoi(context.PostForm("bookId"))
	if err1 != nil || err2 != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	result := agent.BorrowBook(userId, bookId)
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
	result := agent.CountCurrentBorrowedBooks(userId) //历史借阅的图书量
	borrowedBookCount := result.Msg
	result = agent.TotalFineAmount(userId) //总罚款数量
	Totalfine := result.Msg
	result = agent.CountHistoryReservedBooks(userId) //历史预定图书量
	historyReservedBookCount := result.Msg
	result = agent.CountCurrentReservedBooks(userId) //历史预定图书量
	currentReservedBookCount := result.Msg
	result = agent.LastReturnBook(userId) //最后归还的图书
	lastReturnBook := result.Msg
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "Current Borrowed books": borrowedBookCount, "Total Fine Amount": Totalfine, "History Reserve books": historyReservedBookCount, "Current Reserve books": currentReservedBookCount, "Last Return Books": lastReturnBook})
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
	fine := GetMemberFine(agent.DB, userId)
	context.JSON(http.StatusOK, gin.H{"fine": fine})
}

func getMemberHistoryFineHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	fine := GetMemberHistoryFine(agent.DB, userId)
	context.JSON(http.StatusOK, gin.H{"fine": fine})
}

func getMemberPayURLHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	url := agent.GetPayMemberFineURL(userId)
	context.JSON(http.StatusOK, gin.H{"url": url})
}

func getMemberCurrentBorrowCountHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	count := agent.GetMemberCurrentBorrowCount(userId)
	context.JSON(http.StatusOK, gin.H{"count": count})
}

func getMemberCurrentReserveCountHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	count := agent.GetMemberCurrentReserveCount(userId)
	context.JSON(http.StatusOK, gin.H{"count": count})
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

func updateEmailHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	newEmail, ok := context.GetPostForm("newEmail")
	if !ok {
		context.Status(http.StatusBadRequest)
		return
	}
	result := agent.UpdateEmail(userId, newEmail)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

func getMemberHistoryFineListPagesHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	page := agent.GetMemberHistoryFineListPages(userId)
	context.JSON(http.StatusOK, gin.H{"page": page})
}

func getMemberHistoryFineListHandler(context *gin.Context) {
	iUserId, _ := context.Get("userId")
	userId := iUserId.(int)
	page, err := strconv.Atoi(context.PostForm("page"))
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	fineList := agent.GetMemberHistoryFineListByPage(userId, page)
	_, _ = context.Writer.Write(util.JsonEncode(fineList))
}

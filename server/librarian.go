package main

import (
	"github.com/gin-gonic/gin"
	. "lms/services"
	"lms/util"
	"log"
	"net/http"
	"strconv"
)

func addBookHandler(context *gin.Context) {
	isbn := context.PostForm("isbn")
	locationId, err1 := strconv.Atoi(context.PostForm("locationId"))
	categoryId, err2 := strconv.Atoi(context.PostForm("categoryId"))
	count, err3 := strconv.Atoi(context.PostForm("count"))
	if err1 != nil || err2 != nil || err3 != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	book, err := GetMetaDataByISBN(isbn)
	bookIdArr := make([]int, 0)
	if err == nil {
		book.LocationId = locationId
		book.CategoryId = categoryId
		bookIdArr = dbAgent.AddBook(&book, count)
		if len(bookIdArr) > 0 {
			log.Printf("Add Book %v (ISBN:%v) Successfully \n", book.Name, book.Isbn)
		} else {
			log.Printf("Fail To Add Book %v (ISBN:%v)  \n", book.Name, book.Isbn)
		}
	}
	_, _ = context.Writer.Write(util.JsonEncode(bookIdArr))
}

func updateBookHandler(context *gin.Context) {
	bookId, err := strconv.Atoi(context.PostForm("bookId"))
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	isbn := context.PostForm("isbn")
	name := context.PostForm("name")
	author := context.PostForm("author")
	language := context.PostForm("language")
	locationId, err1 := strconv.Atoi(context.PostForm("locationId"))
	categoryId, err2 := strconv.Atoi(context.PostForm("categoryId"))
	if err1 != nil || err2 != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	book := Book{
		Id:       bookId,
		Name:     name,
		Author:   author,
		Isbn:     isbn,
		Language: language,
		LocationId: locationId,
		CategoryId: categoryId,
	}
	result := agent.UpdateBook(&book)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

func deleteBookHandler(context *gin.Context) {
	bookId, err1 := strconv.Atoi(context.PostForm("bookId"))
	state, err2 := strconv.Atoi(context.PostForm("state"))
	if err1 != nil || err2 != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	result := agent.DeleteBook(bookId, BookState(state))
	context.JSON(http.StatusOK, gin.H{"state": result.Status, "msg": result.Msg})
}

func returnBookHandler(context *gin.Context) {
	bookIdString := context.PostForm("bookId")
	bookId, _ := strconv.Atoi(bookIdString)
	result := dbAgent.ReturnBook(bookId)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

func getAllBorrowBooksPagesHandler(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"page": agent.GetBorrowBooksPages()})
}

func getAllBorrowBooksHandler(context *gin.Context) {
	pageStr := context.PostForm("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	books := agent.GetBorrowBooksByPage(page)
	_, _ = context.Writer.Write(util.JsonEncode(books))
}

func getAllMembersPagesHandler(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"page": agent.GetMemberPages()})
}

func getAllMembersHandler(context *gin.Context) {
	pageStr := context.PostForm("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	users := agent.GetMembersByPage(page)
	_, _ = context.Writer.Write(util.JsonEncode(users))
}

func getMembersHasDebtPagesHandler(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"page": agent.GetMembersHasDebtPages()})
}

func getMembersHasDebtHandler(context *gin.Context) {
	pageStr := context.PostForm("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	users := agent.GetMembersHasDebtByPage(page)
	_, _ = context.Writer.Write(util.JsonEncode(users))
}

func deleteMemberHandler(context *gin.Context) {
	userIdStr := context.PostForm("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	result := agent.DeleteMember(userId)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

func addCategoryHandler(context *gin.Context) {
	categoryName, ok := context.GetPostForm("category")
	if !ok {
		context.Status(http.StatusBadRequest)
		return
	}
	result := agent.AddCategory(categoryName)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

func addLocationHandler(context *gin.Context) {
	locationName, ok := context.GetPostForm("location")
	if !ok {
		context.Status(http.StatusBadRequest)
		return
	}
	result := agent.AddLocation(locationName)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

func getMemberCountHandler(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"count": agent.GetMemberCount()})
}

func getBookCountByISBNHandler(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"count": agent.GetBookCountByISBN()})
}

func getBookCountByCopyHandler(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"count": agent.GetBookCountByCopy()})
}

func getCurrentBorrowCountHandler(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"count": agent.GetCurrentBorrowCount()})
}

func getHistoryBorrowCountHandler(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"count": agent.GetHistoryBorrowCount()})
}

func getDamagedBookCountHandler(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"count": agent.GetDamagedBookCount()})
}

func getLostBookCountHandler(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"count": agent.GetLostBookCount()})
}

func getUnpaidFineHandler(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"count": agent.GetUnpaidFine()})
}

func getPaidFineHandler(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"count": agent.GetPaidFine()})
}

func getHistoryFineListPagesHandler(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"page": agent.GetHistoryFineListPages()})
}

func getHistoryFineListHandler(context *gin.Context) {
	page, err := strconv.Atoi(context.PostForm("page"))
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	fineList := agent.GetHistoryFineListByPage(page)
	_, _ = context.Writer.Write(util.JsonEncode(fineList))
}
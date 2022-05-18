package main

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	. "lms/services"
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
	bf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(bf)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(bookIdArr)

	_, _ = context.Writer.Write(bf.Bytes())
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
	bookId, err := strconv.Atoi(context.PostForm("bookId"))
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	result := agent.DeleteBook(bookId)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
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

	bf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(bf)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(books)

	_, _ = context.Writer.Write(bf.Bytes())
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

	bf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(bf)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(users)

	_, _ = context.Writer.Write(bf.Bytes())
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

	bf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(bf)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(users)

	_, _ = context.Writer.Write(bf.Bytes())
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

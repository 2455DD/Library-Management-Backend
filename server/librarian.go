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
	location := context.PostForm("location")
	count, err := strconv.Atoi(context.PostForm("count"))
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}

	book, err := GetMetaDataByISBN(isbn)
	bookIdArr := make([]int, 0)
	if err == nil {
		book.Location = location
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
	address := context.PostForm("address")
	language := context.PostForm("language")
	location := context.PostForm("location")
	book := Book{
		Id:       bookId,
		Name:     name,
		Author:   author,
		Isbn:     isbn,
		Address:  address,
		Language: language,
		Location: location,
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
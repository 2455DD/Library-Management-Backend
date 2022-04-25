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
	if err != nil {
		context.JSON(http.StatusOK, gin.H{"status": AddFailed, "msg": "查询不到ISBN"})
		return
	}
	book.Location = location

	result := dbAgent.AddBook(&book, count)
	if result.Status == UpdateOK {
		log.Printf("Add Book %v (ISBN:%v) Successfully \n", book.Name, book.Isbn)
	} else {
		log.Printf("FAIL TO Add Book %v (ISBN:%v)  \n", book.Name, book.Isbn)
	}
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
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

func getAllBorrowBooksPages(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"page": agent.GetBorrowBooksPages()})
}

func getAllBorrowBooks(context *gin.Context) {
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

func getAllMembersPages(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"page": agent.GetMemberPages()})
}

func getAllMembers(context *gin.Context) {
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
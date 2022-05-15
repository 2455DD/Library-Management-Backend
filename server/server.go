package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/go-ini/ini"
	"io"
	"lms/middlewares"
	. "lms/services"
	"log"
	"net/http"
	"strconv"
)

var port int
var path string
var staticPath string

var agent Agent

func getBooksPagesHandler(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"page": dbAgent.GetBooksPages()})
}

func getBooksHandler(context *gin.Context) {
	pageString := context.Query("page")
	page, err := strconv.Atoi(pageString)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	books := dbAgent.GetBooksByPage(page)
	bf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(bf)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(books)

	_, _ = context.Writer.Write(bf.Bytes())
}

func getBookBarcodeHandler(context *gin.Context) {
	bookId, err := strconv.Atoi(context.Query("bookId"))
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	writer := context.Writer.(io.Writer)
	dbAgent.GetBookBarcode(&writer, bookId)
}

func getMemberBarcodeHandler(context *gin.Context) {
	userId, err := strconv.Atoi(context.Query("userId"))
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	writer := context.Writer.(io.Writer)
	dbAgent.GetMemberBarcode(&writer, userId)
}

func initServer(cfg *ini.File) {
	serverCfg, err := cfg.GetSection("server")
	if err != nil {
		log.Fatal("Fail to load section 'server': ", err)
	}
	port = serverCfg.Key("port").MustInt(80)
	path = serverCfg.Key("path").MustString("")
	staticPath = serverCfg.Key("staticPath").MustString("")
	JiKeAPIkey = serverCfg.Key("JiKeAPIKey").MustString("")

	agent.PayAgent = &payAgent
	agent.DBAgent = &dbAgent
}

func startService() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.LoadHTMLFiles(fmt.Sprintf("%v/index.html", path))
	router.Use(static.Serve("/static", static.LocalFile(staticPath, true)))

	router.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "index.html", nil)
	})

	g1 := router.Group("/")
	g1.Use(middlewares.MemberAuth())
	{
		g1.POST("/getBorrowBooksPages", getBorrowBooksPagesHandler)
		g1.POST("/getBorrowBooks", getBorrowBooksHandler)
		g1.POST("/getReserveBooksPages", getReserveBooksPagesHandler)
		g1.POST("/getReserveBooks", getReserveBooksHandler)
		g1.POST("/reserveBook", reserveBookHandler)
		g1.POST("/cancelReserveBook", cancelReserveBookHandler)
		g1.POST("/getMemberHistoryBorrowTime", getMemberHistoryBorrowTimeHandler)
		g1.POST("/getMemberFine", getMemberFineHandler)
		g1.POST("/getMemberPayURL", getMemberPayURLHandler)
		g1.POST("/updatePassword", updatePasswordHandler)
	}

	g2 := router.Group("/")
	// Librarian Domain, Used After Auth
	g2.Use(middlewares.AdminAuth())
	{
		g2.POST("/getMemberBorrowHistoryByPage", getBorrowBooksByMemberIDHandler)
		g2.POST("/borrowBook", borrowBookHandler)
		g2.POST("/addBook", addBookHandler)
		g2.POST("/register", registerHandler)
		g2.POST("/updateBook", updateBookHandler)
		g2.POST("/deleteBook", deleteBookHandler)
		g2.POST("/returnBook", returnBookHandler)
		g2.POST("/getAllBorrowBooksPages", getAllBorrowBooksPagesHandler)
		g2.POST("/getAllBorrowBooks", getAllBorrowBooksHandler)
		g2.POST("/getAllMembersPages", getAllMembersPagesHandler)
		g2.POST("/getAllMembers", getAllMembersHandler)
		g2.POST("/getMembersHasDebtPages", getMembersHasDebtPagesHandler)
		g2.POST("/getMembersHasDebt", getMembersHasDebtHandler)
	}

	router.POST("/login", loginHandler)
	router.POST("/admin", adminLoginHandler)
	router.GET("/getBooksPages", getBooksPagesHandler)
	router.GET("/getBooks", getBooksHandler)
	router.GET("/getBookBarcode", getBookBarcodeHandler)
	router.GET("/getMemberBarcode", getMemberBarcodeHandler)

	router.StaticFile("/favicon.ico", fmt.Sprintf("%v/favicon.ico", path))

	err := router.Run(":" + strconv.Itoa(port))
	if err != nil {
		fmt.Println(err)
		return
	}
}

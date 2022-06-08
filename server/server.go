package main

import (
	"fmt"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/go-ini/ini"
	"io"
	"lms/middlewares"
	. "lms/services"
	"lms/util"
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

func getBooksPagesByCategoryHandler(context *gin.Context) {
	categoryIdStr := context.Query("categoryId")
	categoryId, err := strconv.Atoi(categoryIdStr)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	context.JSON(http.StatusOK, gin.H{"page": dbAgent.GetBooksPagesByCategory(categoryId)})
}

func getBooksPagesByLocationHandler(context *gin.Context) {
	locationIdStr := context.Query("locationId")
	locationId, err := strconv.Atoi(locationIdStr)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	context.JSON(http.StatusOK, gin.H{"page": dbAgent.GetBooksPagesByLocation(locationId)})
}

func getBooksHandler(context *gin.Context) {
	pageString := context.Query("page")
	page, err := strconv.Atoi(pageString)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	books := dbAgent.GetBooksByPage(page)
	_, _ = context.Writer.Write(util.JsonEncode(books))
}

func getBooksByCategoryHandler(context *gin.Context) {
	pageString := context.Query("page")
	page, err := strconv.Atoi(pageString)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	categoryIdStr := context.Query("categoryId")
	categoryId, err := strconv.Atoi(categoryIdStr)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	books := dbAgent.GetBooksByCategory(page, categoryId)
	_, _ = context.Writer.Write(util.JsonEncode(books))
}

func getBooksByLocationHandler(context *gin.Context) {
	pageString := context.Query("page")
	page, err := strconv.Atoi(pageString)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	locationIdStr := context.Query("locationId")
	locationId, err := strconv.Atoi(locationIdStr)
	if err != nil {
		context.Status(http.StatusBadRequest)
		return
	}
	books := dbAgent.GetBooksByCategory(page, locationId)
	_, _ = context.Writer.Write(util.JsonEncode(books))
}

func getCategoriesHandler(context *gin.Context) {
	categories := dbAgent.GetCategories()
	_, _ = context.Writer.Write(util.JsonEncode(categories))
}

func getLocationsHandler(context *gin.Context) {
	locations := dbAgent.GetLocations()
	_, _ = context.Writer.Write(util.JsonEncode(locations))
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
		g1.POST("/getMemberHistoryFine", getMemberHistoryFineHandler)
		g1.POST("/getMemberPayURL", getMemberPayURLHandler)
		g1.POST("/getMemberCurrentBorrowCount", getMemberCurrentBorrowCountHandler)
		g1.POST("/getMemberCurrentReserveCount", getMemberCurrentReserveCountHandler)
		g1.POST("/updatePassword", updatePasswordHandler)
		g1.POST("/updateEmail", updateEmailHandler)
		g1.POST("/getMemberHistoryFineListPages", getMemberHistoryFineListPagesHandler)
		g1.POST("/getMemberHistoryFineList", getMemberHistoryFineListHandler)
	}

	g2 := router.Group("/")
	g2.Use(middlewares.AdminAuth())
	{
		g2.POST("/addBook", addBookHandler)
		g2.POST("/register", registerHandler)
		g2.POST("/updateBook", updateBookHandler)
		g2.POST("/deleteBook", deleteBookHandler)
		g2.POST("/borrowBook", borrowBookHandler)
		g2.POST("/returnBook", returnBookHandler)
		g2.POST("/getAllBorrowBooksPages", getAllBorrowBooksPagesHandler)
		g2.POST("/getAllBorrowBooks", getAllBorrowBooksHandler)
		g2.POST("/getAllMembersPages", getAllMembersPagesHandler)
		g2.POST("/getAllMembers", getAllMembersHandler)
		g2.POST("/getMembersHasDebtPages", getMembersHasDebtPagesHandler)
		g2.POST("/getMembersHasDebt", getMembersHasDebtHandler)
		g2.POST("/deleteMember", deleteMemberHandler)
		g2.POST("/addCategory", addCategoryHandler)
		g2.POST("/addLocation", addLocationHandler)
		g2.POST("/getMemberCount", getMemberCountHandler)
		g2.POST("/getBookCountByISBN", getBookCountByISBNHandler)
		g2.POST("/getBookCountByCopy", getBookCountByCopyHandler)
		g2.POST("/getCurrentBorrowCount", getCurrentBorrowCountHandler)
		g2.POST("/getHistoryBorrowCount", getHistoryBorrowCountHandler)
		g2.POST("/getDamagedBookCount", getDamagedBookCountHandler)
		g2.POST("/getLostBookCount", getLostBookCountHandler)
		g2.POST("/getUnpaidFine", getUnpaidFineHandler)
		g2.POST("/getPaidFine", getPaidFineHandler)
		g2.POST("/getHistoryFineListPages", getHistoryFineListPagesHandler)
		g2.POST("/getHistoryFineList", getHistoryFineListHandler)
		g2.POST("/getMemberBorrowHistoryPages", getMemberBorrowHistoryPagesHandler)
		g2.POST("/getMemberBorrowHistory", getMemberBorrowHistoryHandler)
		g2.POST("/getMemberReturnHistoryPages", getMemberReturnHistoryPagesHandler)
		g2.POST("/getMemberReturnHistory", getMemberReturnHistoryHandler)
		g2.POST("/getMemberReserveHistoryPages", getMemberReserveHistoryPagesHandler)
		g2.POST("/getMemberReserveHistory", getMemberReserveHistoryHandler)
		g2.POST("/getMemberFineHistoryPages", getMemberFineHistoryPagesHandler)
		g2.POST("/getMemberFineHistory", getMemberFineHistoryHandler)
	}

	router.POST("/login", loginHandler)
	router.POST("/admin", adminLoginHandler)
	router.GET("/getBooksPages", getBooksPagesHandler)
	router.GET("/getBooks", getBooksHandler)
	router.GET("/getCategories", getCategoriesHandler)
	router.GET("/getLocations", getLocationsHandler)
	router.GET("/getBooksPagesByCategory", getBooksPagesByCategoryHandler)
	router.GET("/getBooksByCategory", getBooksByCategoryHandler)
	router.GET("/getBooksPagesByLocation", getBooksPagesByLocationHandler)
	router.GET("/getBooksByLocation", getBooksByLocationHandler)
	router.GET("/getBookBarcode", getBookBarcodeHandler)
	router.GET("/getMemberBarcode", getMemberBarcodeHandler)

	router.StaticFile("/favicon.ico", fmt.Sprintf("%v/favicon.ico", path))

	err := router.Run(":" + strconv.Itoa(port))
	if err != nil {
		fmt.Println(err)
		return
	}
}

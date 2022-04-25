package services

import (
	"encoding/json"
	"errors"
	"fmt"
	goISBN "github.com/abx123/go-isbn"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// 极客API,用于获得中文书籍ISBN信息
var JiKeAPIkey string

type bookRetrieveResult struct {
	Name        string `json:"name"`        //书名
	Author      string `json:"author"`      //作者
	AuthorIntro string `json:"authorIntro"` //作者简介
	PhotoUrl    string `json:"photoUrl"`    //图片封面
	Publishing  string `json:"publishing"`  //出版社
	Published   string `json:"published"`   //出版时间
	Description string `json:"description"` //图书简介
	DoubanScore int    `json:"doubanScore"` //豆瓣评分
}

type bookRetrieveHTTPResult struct {
	Ret  int                `json:"ret"`
	Msg  string             `json:"msg"`
	Data bookRetrieveResult `json:"Data"`
}

var myClient = &http.Client{
	Timeout: 10 * time.Second,
	//Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
}

func getJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		log.Printf("HTTP Request from %v Failure:%v\n", url, err.Error())
		return err
	}
	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(target)
	if err != nil {
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		bodyString := string(bodyBytes)
		log.Printf("Get non-json reply from %v\nStatusCode:%v Body:%v", url, r.StatusCode, bodyString)
	}
	return err
}

func GetMetaDataByISBN(isbn string) (bookInfo Book, err error) {
	bookInfo = Book{}
	isbnRetriever := goISBN.NewGoISBN(goISBN.DEFAULT_PROVIDERS)
	rawBookInfo, err := isbnRetriever.Get(isbn)
	if err != nil {
		log.Printf("ISBN %v not found in Google Books and Open Library, Checking Jike\n", isbn)
		var resp bookRetrieveHTTPResult
		// FIXME: Always Get 403
		err := getJson(fmt.Sprintf("https://api.jike.xyz/situ/book/isbn/%v?apikey=%v", isbn, JiKeAPIkey), &resp)
		if err != nil {
			bookInfo.Isbn = isbn
			bookInfo.Name = "Unknown"
			bookInfo.Author = "Unknown"
			bookInfo.Language = "Unknown"
			return bookInfo, err
		}
		if resp.Ret != 0 {
			return bookInfo, errors.New("book cannot be retrieved, no result")
		}
		bookInfo.Name = resp.Data.Name
		bookInfo.Author = resp.Data.Author
		bookInfo.Language = "Chinese"
	} else {
		bookInfo.Isbn = isbn
		var authors string
		for _, subAuthor := range rawBookInfo.Authors {
			authors += fmt.Sprintf("%v,", subAuthor)
		}
		bookInfo.Author = authors
		bookInfo.Language = rawBookInfo.Language
		bookInfo.Name = rawBookInfo.Title
	}
	return bookInfo, nil
}

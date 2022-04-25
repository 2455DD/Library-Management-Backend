package services

import (
	"github.com/smartwalle/alipay/v3"
	. "gorm.io/gorm"
)

type StatusCode int

type DBAgent struct {
	DB *DB
}

type PayAgent struct {
	PayClient *alipay.Client
}

type Agent struct {
	*DBAgent
	*PayAgent
}

type Book struct {
	Id       int    `column:"id" gorm:"primaryKey"`
	Name     string `column:"name"`
	Author   string `column:"author"`
	Isbn     string `column:"isbn"`
	Address  string `column:"address"`
	Language string `column:"language"`
	Location string `column:"location"`
}

type ReserveBook struct {
	Id		  int		`gorm:"column:id;primaryKey"`
	BookId    int		`gorm:"column:book_id"`
	UserId    int		`gorm:"column:user_id"`
	StartTime string	`gorm:"column:start_time"`
	EndTime   string	`gorm:"column:end_time"`
}

type BorrowBook struct {
	Id		  int		`gorm:"column:id;primaryKey"`
	BookId    int		`gorm:"column:book_id"`
	UserId    int		`gorm:"column:user_id"`
	StartTime string	`gorm:"column:start_time"`
	EndTime   string	`gorm:"column:end_time"`
}

type ReserveBookStatus struct {
	Book
	StartTime string
	EndTime   string
}

type BorrowBookStatus struct {
	Book
	StartTime string
	EndTime   string
	Fine      int
}

type Pay struct {
	Id       int	`gorm:"column:id;primaryKey"`
	UserId   int    `gorm:"column:user_id"`
	Amount   int	`gorm:"column:amount"`
	Done     int    `gorm:"column:done"`
}

type StatusResult struct {
	Code   int
	Msg    string
	Status StatusCode
}

const (
	OK StatusCode = iota
	LoginIdNotExist
	LoginIdOrPasswordError
	LoginOK
	RegisterFailed
	RegisterOK
	ReserveFailed
	ReserveOK
	CancelReserveFailed
	CancelReserveOK
	BorrowFailed
	BorrowOK
	ReturnFailed
	ReturnOK
	AddFailed
	AddOK
	UpdateFailed
	UpdateOK
	DeleteFailed
	DeleteOK
)

var (
	MediaPath string
)

func (book *Book) TableName() string {
	return "book"
}

func (reserveBook *ReserveBook) TableName() string {
	return "reserve"
}

func (borrowBook *BorrowBook) TableName() string {
	return "borrow"
}

func (pay *Pay) TableName() string{
	return "pay"
}

func (agent *DBAgent) GetBooksPages() int64 {
	var count int64
	if err := agent.DB.Table("book").Count(&count).Error; err != nil {
		return 0
	}
	return count / 10 + 1
}

func (agent *DBAgent) GetBooksByPage(page int) []Book {
	books := make([]Book, 0)
	agent.DB.Offset((page - 1) * 10).Limit(10).Find(&books)
	return books
}
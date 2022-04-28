package services

import (
	"errors"
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

type BookData struct {
	Book
	Status BookStatus
}

type BookStatus int

const (
	Idle BookStatus = iota
	Reserved
	Borrowed
)

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
	UpdatePasswordFailed
	UpdatePasswordOK
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

func (agent *DBAgent) GetBooksByPage(page int) []BookData {
	books := make([]Book, 0)
	bookDataArr := make([]BookData, 0)
	agent.DB.Offset((page - 1) * 10).Limit(10).Find(&books)
	for _, book := range books {
		bookData := BookData{
			Book:   book,
			Status: Idle,
		}
		if err := agent.DB.Where("book_id = ? and end_time is null", book.Id).Last(&ReserveBook{}).Error; err == nil {
			bookData.Status = Reserved
		}
		if err := agent.DB.Where("book_id = ? and end_time is null", book.Id).Last(&BorrowBook{}).Error; err == nil {
			bookData.Status = Borrowed
		}
		bookDataArr = append(bookDataArr, bookData)
	}
	return bookDataArr
}

func (agent *DBAgent) UpdatePassword(userId int, oldPassword string, newPassword string) StatusResult {
	result := StatusResult{}
	err := agent.DB.Transaction(func(tx *DB) error {
		user := User{}
		if err := tx.First(&user, userId).Error; err != nil {
			return err
		}
		if user.Password != oldPassword {
			return errors.New("密码不正确")
		}
		if err := tx.Model(&user).Update("password", newPassword).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		result.Status = UpdatePasswordFailed
		result.Msg = "修改密码失败"
		return result
	}
	result.Status = UpdateOK
	result.Msg = "修改密码成功"
	return result
}
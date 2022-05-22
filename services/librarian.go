package services

import (
	"gorm.io/gorm"
	"lms/util"
	"time"
)

type Librarian struct {
	UserId   int    `db:"id" gorm:"column:id;primaryKey"`
	Username string `db:"username" gorm:"column:username"`
	Password string `db:"password" gorm:"column:password"`
}

type BorrowData struct {
	BorrowBookStatus
	UserId int
}

type ReserveData struct {
	ReserveBookStatus
	UserId int
}

type UserData struct {
	Id       int
	Username string
	Email    string
	Debt     int
}

func (librarian Librarian) TableName() string {
	return "admin"
}

func (agent *DBAgent) AuthenticateLibrarian(userId int, password string) StatusResult {
	result := StatusResult{}
	librarian := &Librarian{}
	tx := agent.DB.First(&librarian, userId)
	if tx.Error != nil {
		result.Status = LoginIdNotExist
		result.Msg = "不存在此用户"
		return result
	}
	if password != librarian.Password {
		result.Status = LoginIdOrPasswordError
		result.Msg = "密码错误"
		return result
	}
	result.Status = LoginOK
	result.Msg = "登录成功"
	return result
}

func (agent *DBAgent) RegisterMember(user *User) StatusResult {
	result := StatusResult{}
	err := agent.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("username", "password", "email").Create(user).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		result.Status = RegisterFailed
		result.Msg = "注册失败"
		return result
	}
	result.Status = RegisterOK
	result.Msg = "注册成功"
	return result
}

func (agent *DBAgent) AddBook(book *Book, count int) []int {
	bookIdArr := make([]int, 0)
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < count; i++ {
			res := tx.Omit("id").Create(book)
			if err := res.Error; err != nil {
				return err
			}
			bookIdArr = append(bookIdArr, book.Id)
			book.Id = 0
		}
		return nil
	})
	return bookIdArr
}

func (agent *DBAgent) UpdateBook(book *Book) StatusResult {
	result := StatusResult{}
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", book.Id).Model(book).Omit("id").Updates(book).Error; err != nil {
			result.Status = UpdateFailed
			result.Msg = "更新图书信息失败"
			return err
		}
		result.Status = UpdateOK
		result.Msg = "更新图书信息成功"
		return nil
	})
	return result
}

func (agent *DBAgent) DeleteBook(bookId int) StatusResult {
	result := StatusResult{}
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		borrowBook := BorrowBook{}
		if err := tx.Where("book_id = ?", bookId).Last(&borrowBook).Error; err == nil {
			if borrowBook.EndTime == "" {
				result.Status = DeleteFailed
				result.Msg = "删除图书失败"
				return nil
			}
		}
		reserveBook := ReserveBook{}
		if err := tx.Where("book_id = ?", bookId).Last(&reserveBook).Error; err == nil {
			if reserveBook.EndTime == "" {
				result.Status = DeleteFailed
				result.Msg = "删除图书失败"
				return nil
			}
		}
		if err := tx.Delete(&Book{}, bookId).Error; err != nil {
			result.Status = DeleteFailed
			result.Msg = "删除图书失败"
			return err
		}
		result.Status = DeleteOK
		result.Msg = "删除图书成功"
		return nil
	})
	return result
}

func (agent *DBAgent) GetBorrowBooksPages() int64 {
	var count int64
	if err := agent.DB.Table("borrow").Count(&count).Error; err != nil {
		return 0
	}
	return count/10 + 1
}

func (agent *DBAgent) GetBorrowBooksByPage(page int) []BorrowData {
	borrowDataArr := make([]BorrowData, 0)
	borrowBooks := make([]BorrowBook, 0)
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		tx.Offset((page - 1) * 10).Limit(10).Find(&borrowBooks)
		for _, borrowBook := range borrowBooks {
			book := Book{}
			if err := tx.First(&book, borrowBook.BookId).Error; err == nil {
				status := BorrowBookStatus{}
				status.Book = book
				status.StartTime = borrowBook.StartTime
				status.EndTime = borrowBook.EndTime
				deadline := util.StringToTime(borrowBook.StartTime).Add(time.Hour * 240)
				status.Deadline = deadline.Format(util.GormTimeFormat)
				status.Fine = CalculateFine(status)

				data := BorrowData{}
				data.BorrowBookStatus = status
				data.UserId = borrowBook.UserId

				borrowDataArr = append(borrowDataArr, data)
			}
		}
		return nil
	})
	return borrowDataArr
}

func (agent *DBAgent) GetMemberActiveBorrowHistoryByPage(page int, userID int) []BorrowData {
	borrowDataArr := make([]BorrowData, 0)
	borrowBooks := make([]BorrowBook, 0)
	_ = agent.DB.Transaction(
		func(tx *gorm.DB) error {
			tx.Offset((page-1)*10).Limit(10).
				Where("user_id = ? AND NOW() <= DATE_ADD(borrow.createtime,INTERVAL 10 DAY)", userID).Find(&borrowBooks)
			for _, borrowBook := range borrowBooks {
				book := Book{}
				if err := tx.First(&book, borrowBook.BookId).Error; err == nil {
					status := BorrowBookStatus{}
					status.Book = book
					status.StartTime = borrowBook.StartTime
					status.EndTime = borrowBook.EndTime
					deadline := util.StringToTime(borrowBook.StartTime).Add(time.Hour * 240)
					status.Deadline = deadline.Format(util.GormTimeFormat)
					status.Fine = CalculateFine(status)

					data := BorrowData{}
					data.BorrowBookStatus = status
					data.UserId = borrowBook.UserId

					borrowDataArr = append(borrowDataArr, data)
				}
			}
			return nil
		})
	return borrowDataArr
}

func (agent *DBAgent) GetMemberOverdueHistoryByPage(page int, userID int) []BorrowData {
	borrowDataArr := make([]BorrowData, 0)
	borrowBooks := make([]BorrowBook, 0)
	_ = agent.DB.Transaction(
		func(tx *gorm.DB) error {
			tx.Offset((page-1)*10).Limit(10).
				Where("user_id = ? AND NOW() >= DATE_ADD(borrow.createtime,INTERVAL 10 DAY)  AND endtime IS NULL",
					userID).Find(&borrowBooks)
			for _, borrowBook := range borrowBooks {
				book := Book{}
				if err := tx.First(&book, borrowBook.BookId).Error; err == nil {
					status := BorrowBookStatus{}
					status.Book = book
					status.StartTime = borrowBook.StartTime
					status.EndTime = borrowBook.EndTime
					deadline := util.StringToTime(borrowBook.StartTime).Add(time.Hour * 240)
					status.Deadline = deadline.Format(util.GormTimeFormat)
					status.Fine = CalculateFine(status)

					data := BorrowData{}
					data.BorrowBookStatus = status
					data.UserId = borrowBook.UserId

					borrowDataArr = append(borrowDataArr, data)
				}
			}
			return nil
		})
	return borrowDataArr
}

func (agent *DBAgent) GetMemberReserveHistoryByPage(page int, userID int) []ReserveData {

	reserveData := make([]ReserveBook, 0)
	reserveBookData := make([]ReserveData, 0)
	_ = agent.DB.Transaction(
		func(tx *gorm.DB) error {
			tx.Offset((page-1)*10).Limit(10).
				Where("user_id = ? AND endtime IS NULL)", userID).Find(&reserveData)
			for _, reserveEntry := range reserveData {
				book := Book{}
				if err := tx.First(&book, reserveEntry.BookId).Error; err == nil {
					status := ReserveBookStatus{}
					status.Book = book
					status.StartTime = reserveEntry.StartTime
					status.EndTime = reserveEntry.EndTime

					data := ReserveData{}
					data.ReserveBookStatus = status
					data.UserId = reserveEntry.UserId

					reserveBookData = append(reserveBookData, data)
				}
			}
			return nil
		})
	return reserveBookData
}

func (agent DBAgent) GetMemberReturnHistoryByPage(page int, userID int) []BorrowData {
	// FIXME:WARNING,UNTESTED
	borrowDataArr := make([]BorrowData, 0)
	borrowBooks := make([]BorrowBook, 0)
	_ = agent.DB.Transaction(
		func(tx *gorm.DB) error {
			tx.Offset((page-1)*10).Limit(10).
				Where("user_id = ? AND endtime IS NOT NULL",
					userID).Find(&borrowBooks)
			for _, borrowBook := range borrowBooks {
				book := Book{}
				if err := tx.First(&book, borrowBook.BookId).Error; err == nil {
					status := BorrowBookStatus{}
					status.Book = book
					status.StartTime = borrowBook.StartTime
					status.EndTime = borrowBook.EndTime
					deadline := util.StringToTime(borrowBook.StartTime).Add(time.Hour * 240)
					status.Deadline = deadline.Format(util.GormTimeFormat)
					status.Fine = CalculateFine(status)

					data := BorrowData{}
					data.BorrowBookStatus = status
					data.UserId = borrowBook.UserId

					borrowDataArr = append(borrowDataArr, data)
				}
			}
			return nil
		})
	return borrowDataArr
}

func (agent *DBAgent) GetMemberFineHistoryByPage(page int, userID int) {
	//TODO: DO THE FUCKING ME!
}

func (agent *DBAgent) GetMemberPages() int64 {
	var count int64
	if err := agent.DB.Table("user").Count(&count).Error; err != nil {
		return 0
	}
	return count/10 + 1
}

func (agent *Agent) GetMembersByPage(page int) []UserData {
	userArr := make([]UserData, 0)
	users := make([]User, 0)
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		tx.Offset((page - 1) * 10).Limit(10).Find(&users)
		for _, user := range users {
			agent.GetMemberFine(user.UserId)
			userData := &UserData{
				Id:       user.UserId,
				Username: user.Username,
				Email:    user.Email,
				Debt:     user.Debt,
			}
			userArr = append(userArr, *userData)
		}
		return nil
	})
	return userArr
}

func (agent *DBAgent) GetMembersHasDebtPages() int64 {
	var count int64
	if err := agent.DB.Table("user").Where("debt > 0").Count(&count).Error; err != nil {
		return 0
	}
	return count/10 + 1
}

func (agent *Agent) GetMembersHasDebtByPage(page int) []UserData {
	userArr := make([]UserData, 0)
	users := make([]User, 0)
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("debt > 0").Offset((page - 1) * 10).Limit(10).Find(&users)
		for _, user := range users {
			agent.GetMemberFine(user.UserId)
			userData := &UserData{
				Id:       user.UserId,
				Username: user.Username,
				Email:    user.Email,
				Debt:     user.Debt,
			}
			userArr = append(userArr, *userData)
		}
		return nil
	})
	return userArr
}

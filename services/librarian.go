package services

import (
	"gorm.io/gorm"
)

type Librarian struct {
	UserID   int    `db:"id" gorm:"column:id;primaryKey"`
	Username string `db:"username" gorm:"column:username"`
	Password string `db:"password" gorm:"column:password"`
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

func (agent *DBAgent) AddBook(book *Book, count int) StatusResult {
	result := StatusResult{}
	err := agent.DB.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < count; i++ {
			if err := tx.Omit("id").Create(book).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		result.Status = AddFailed
		result.Msg = "添加图书失败"
	} else {
		result.Status = AddOK
		result.Msg = "添加图书成功"
	}
	return result
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
	return count / 10 + 1
}

func (agent *DBAgent) GetBorrowBooksByPage(page int) []BorrowBookStatus {
	statusArr := make([]BorrowBookStatus, 0)
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
				status.Fine = CalculateFine(status)
				statusArr = append(statusArr, status)
			}
		}
		return nil
	})
	return statusArr
}

func (agent *DBAgent) GetMemberPages() int64 {
	var count int64
	if err := agent.DB.Table("user").Count(&count).Error; err != nil {
		return 0
	}
	return count / 10 + 1
}

func (agent *DBAgent) GetMembersByPage(page int) []UserData {
	userArr := make([]UserData, 0)
	users := make([]User, 0)
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		tx.Offset((page - 1) * 10).Limit(10).Find(&users)
		for _, user := range users {
			userData := &UserData{
				Id:       user.UserID,
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
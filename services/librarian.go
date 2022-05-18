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
	return count / 10 + 1
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
				status.BookMetaData = agent.getBookData(&book)
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

func (agent *DBAgent) GetMemberPages() int64 {
	var count int64
	if err := agent.DB.Table("user").Count(&count).Error; err != nil {
		return 0
	}
	return count / 10 + 1
}

func (agent *Agent) GetMembersByPage(page int) []UserData {
	userArr := make([]UserData, 0)
	users := make([]User, 0)
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		tx.Offset((page - 1) * 10).Limit(10).Find(&users)
		for _, user := range users {
			GetMemberFine(tx, user.UserId)
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
	return count / 10 + 1
}

func (agent *Agent) GetMembersHasDebtByPage(page int) []UserData {
	userArr := make([]UserData, 0)
	users := make([]User, 0)
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("debt > 0").Offset((page - 1) * 10).Limit(10).Find(&users)
		for _, user := range users {
			GetMemberFine(tx, user.UserId)
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

func (agent *Agent) DeleteMember(userId int) StatusResult {
	result := StatusResult{}
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		borrows := make([]BorrowBook, 0)
		reserves := make([]ReserveBook, 0)
		ok := true
		if err := tx.Table("borrow").Where("user_id = ? and end_time is null", userId).Find(&borrows).Error; err != nil {
			ok = false
		}
		if err := tx.Table("reserve").Where("user_id = ? and end_time is null", userId).Find(&reserves).Error; err != nil {
			ok = false
		}
		user := &User{}
		GetMemberFine(tx, userId)
		if err := tx.First(&user, userId).Error; err != nil || user.Debt > 0 {
			ok = false
		}
		if !ok || len(borrows) > 0 || len(reserves) > 0 {
			result.Status = DeleteUserFailed
			result.Msg = "删除用户失败"
			return nil
		}
		if err := tx.Delete(&User{}, userId).Error; err != nil {
			result.Status = DeleteUserFailed
			result.Msg = "删除用户失败"
			return nil
		} else {
			result.Status = DeleteUserOK
			result.Msg = "删除用户成功"
		}
		return nil
	})
	return result
}

func (agent *Agent) AddCategory(name string) StatusResult {
	result := StatusResult{}
	err := agent.DB.Transaction(func(tx *gorm.DB) error {
		category := Category{}
		category.Name = name
		if err := tx.Select("name").Create(&category).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		result.Status = AddCategoryFailed
		result.Msg = "添加Category失败"
	} else {
		result.Status = AddCategoryOK
		result.Msg = "添加Category成功"
	}
	return result
}

func (agent *Agent) AddLocation(name string) StatusResult {
	result := StatusResult{}
	err := agent.DB.Transaction(func(tx *gorm.DB) error {
		location := Location{}
		location.Name = name
		if err := tx.Select("name").Create(&location).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		result.Status = AddLocationFailed
		result.Msg = "添加Location失败"
	} else {
		result.Status = AddLocationOK
		result.Msg = "添加Location成功"
	}
	return result
}
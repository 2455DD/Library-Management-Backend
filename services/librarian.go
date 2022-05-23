package services

import (
	"errors"
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
	book.State = Idle
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
	err := agent.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", book.Id).Model(book).Omit("id").Updates(book).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		result.Status = UpdateFailed
		result.Msg = "更新图书信息失败"
	} else {
		result.Status = UpdateOK
		result.Msg = "更新图书信息成功"
	}
	return result
}

func (agent *DBAgent) DeleteBook(bookId int, state BookState) StatusResult {
	result := StatusResult{}
	if state != Damaged && state != Lost {
		result.Status = DeleteFailed
		result.Msg = "删除图书失败"
		return result
	}
	err := agent.DB.Transaction(func(tx *gorm.DB) error {
		book := Book{}
		if err := tx.First(&book, bookId).Error; err != nil {
			return err
		}
		if err := tx.Model(&book).Update("state", state).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		result.Status = DeleteFailed
		result.Msg = "删除图书失败"
	} else {
		result.Status = DeleteOK
		result.Msg = "删除图书成功"
	}
	return result
}

func (agent *DBAgent) GetBorrowBooksPages() int64 {
	var count int64
	if err := agent.DB.Table("borrow").Count(&count).Error; err != nil {
		return 0
	}
	return (count - 1) / itemsPerPage + 1
}

func (agent *DBAgent) GetBorrowBooksByPage(page int) []BorrowData {
	borrowDataArr := make([]BorrowData, 0)
	borrowBooks := make([]BorrowBook, 0)
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		tx.Offset((page - 1) * itemsPerPage).Limit(itemsPerPage).Find(&borrowBooks)
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
	if err := agent.DB.Table("user").Where("state = ?", Available).Count(&count).Error; err != nil {
		return 0
	}
	return (count - 1) / itemsPerPage + 1
}

func (agent *Agent) GetMembersByPage(page int) []UserData {
	userArr := make([]UserData, 0)
	users := make([]User, 0)
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("state = ?", Available).Offset((page - 1) * itemsPerPage).Limit(itemsPerPage).Find(&users)
		for _, user := range users {
			fine := GetMemberFine(tx, user.UserId)
			userData := &UserData{
				Id:       user.UserId,
				Username: user.Username,
				Email:    user.Email,
				Debt:     fine,
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
	return (count - 1) / itemsPerPage + 1
}

func (agent *Agent) GetMembersHasDebtByPage(page int) []UserData {
	userArr := make([]UserData, 0)
	users := make([]User, 0)
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("debt > 0").Offset((page - 1) * itemsPerPage).Limit(itemsPerPage).Find(&users)
		for _, user := range users {
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
	err := agent.DB.Transaction(func(tx *gorm.DB) error {
		GetMemberFine(tx, userId)
		user := &User{}
		if err := tx.First(&user, userId).Error; err != nil {
			return err
		}
		if user.Debt > 0 {
			return errors.New("删除用户失败，用户还有罚款")
		}
		borrows := make([]BorrowBook, 0)
		reserves := make([]ReserveBook, 0)
		if err := tx.Table("borrow").Where("user_id = ? and end_time is null", userId).Find(&borrows).Error; err == nil {
			if len(borrows) > 0 {
				return errors.New("删除用户失败，用户还有未归还的图书")
			}
		}
		if err := tx.Table("reserve").Where("user_id = ? and end_time is null", userId).Find(&reserves).Error; err == nil {
			if len(reserves) > 0 {
				return errors.New("删除用户失败，用户还有预约的图书")
			}
		}
		if err := tx.Model(&user).Update("state", Unavailable).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		result.Status = DeleteUserFailed
		result.Msg = err.Error()
	} else {
		result.Status = DeleteUserOK
		result.Msg = "删除用户成功"
	}
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

func (agent *DBAgent) GetMemberCount() int {
	var count int64
	if err := agent.DB.Table("user").Where("state = ?", Available).Count(&count).Error; err != nil {
		return 0
	}
	return int(count)
}

func (agent *DBAgent) GetBookCountByISBN() int {
	var count int64
	if err := agent.DB.Table("book").Where("state < ?", Damaged).Group("isbn").Count(&count).Error; err != nil {
		return 0
	}
	return int(count)
}

func (agent *DBAgent) GetBookCountByCopy() int {
	var count int64
	if err := agent.DB.Table("book").Where("state < ?", Damaged).Count(&count).Error; err != nil {
		return 0
	}
	return int(count)
}

func (agent *DBAgent) GetCurrentBorrowCount() int {
	var count int64
	if err := agent.DB.Table("book").Where("state = ?", Borrowed).Count(&count).Error; err != nil {
		return 0
	}
	return int(count)
}

func (agent *DBAgent) GetHistoryBorrowCount() int {
	var count int64
	if err := agent.DB.Table("borrow").Count(&count).Error; err != nil {
		return 0
	}
	return int(count)
}

func (agent *DBAgent) GetDamagedBookCount() int {
	var count int64
	if err := agent.DB.Table("book").Where("state = ?", Damaged).Count(&count).Error; err != nil {
		return 0
	}
	return int(count)
}

func (agent *DBAgent) GetLostBookCount() int {
	var count int64
	if err := agent.DB.Table("book").Where("state = ?", Lost).Count(&count).Error; err != nil {
		return 0
	}
	return int(count)
}

func (agent *DBAgent) GetUnpaidFine() int {
	unpaidFine := 0
	borrowBooks := make([]BorrowBook, 0)
	pays := make([]Pay, 0)
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		tx.Find(&borrowBooks)
		tx.Where("done = ?", 1).Find(&pays)
		for _, borrowBook := range borrowBooks {
			status := BorrowBookStatus{}
			status.StartTime = borrowBook.StartTime
			status.EndTime = borrowBook.EndTime
			unpaidFine += CalculateFine(status)
		}
		for _, pay := range pays {
			unpaidFine -= pay.Amount
		}
		return nil
	})
	return unpaidFine
}

func (agent *DBAgent) GetPaidFine() int {
	paidFine := 0
	pays := make([]Pay, 0)
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("done = ?", 1).Find(&pays)
		for _, pay := range pays {
			paidFine += pay.Amount
		}
		return nil
	})
	return paidFine
}

func (agent *DBAgent) GetHistoryFineListPages() int64 {
	var count int64
	if err := agent.DB.Table("pay").Count(&count).Error; err != nil {
		return 0
	}
	return (count - 1) / itemsPerPage + 1
}

func (agent *DBAgent) GetHistoryFineListByPage(page int) []FineData {
	fineDataArr := make([]FineData, 0)
	pays := make([]Pay, 0)
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		tx.Offset((page - 1) * itemsPerPage).Limit(itemsPerPage).Find(&pays)
		for _, pay := range pays {
			fineData := FineData{
				Fine: pay.Amount,
				Done: pay.Done,
			}
			fineDataArr = append(fineDataArr, fineData)
		}
		return nil
	})
	return fineDataArr
}

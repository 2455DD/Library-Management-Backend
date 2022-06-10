package services

import (
	"errors"
	"fmt"
	"github.com/smartwalle/alipay/v3"
	"gorm.io/gorm"
	"lms/util"
	"log"
	"math"
	"strconv"
	"time"
)

type User struct {
	UserId   int         `gorm:"column:id;primaryKey"`
	Username string      `gorm:"column:username"`
	Password string      `gorm:"column:password"`
	Email    string      `gorm:"column:email"`
	Debt     int         `gorm:"column:debt"`
	State    MemberState `gorm:"column:state"`
}

func (user User) TableName() string {
	return "user"
}

type LoginInterface func(userId int, password string) StatusResult

func (agent *DBAgent) AuthenticateUser(userId int, password string) StatusResult {
	result := StatusResult{}
	user := &User{}
	tx := agent.DB.First(&user, userId)
	if tx.Error != nil || user.State == Unavailable {
		result.Status = LoginIdNotExist
		result.Msg = "不存在此用户"
		return result
	}
	if password != user.Password {
		result.Status = LoginIdOrPasswordError
		result.Msg = "密码错误"
		return result
	}
	result.Status = LoginOK
	result.Msg = "用户登录成功"
	return result
}

func (agent *DBAgent) GetMemberBorrowBooksPages(userId int) int64 {
	var count int64
	if err := agent.DB.Table("borrow").Where("user_id = ?", userId).Count(&count).Error; err != nil {
		return 0
	}
	return (count - 1) / itemsPerPage + 1
}

func CalculateDayDiff(startTime time.Time, endTime time.Time) int {
	days := 0
	hours := endTime.Sub(startTime).Hours()
	days += int(math.Ceil(hours / 24))
	return days
}

func CalculateFine(status BorrowBookStatus) int {
	fine := 0
	start := util.StringToTime(status.StartTime)
	end := time.Now()
	if status.EndTime != "" {
		end = util.StringToTime(status.EndTime)
	}
	days := CalculateDayDiff(start, end) - 10
	if days > 0 {
		fine += days
	}
	return fine
}

func (agent *DBAgent) GetMemberBorrowBooks(userId int, page int) []BorrowBookStatus {
	borrowBooks := make([]BorrowBook, 0)
	statusArr := make([]BorrowBookStatus, 0)
	agent.DB.Where("user_id = ?", userId).Find(&borrowBooks).Offset((page - 1) * itemsPerPage).Limit(itemsPerPage)
	for _, borrowBook := range borrowBooks {
		book := Book{}
		if err := agent.DB.First(&book, borrowBook.BookId).Error; err == nil {
			status := BorrowBookStatus{}
			status.BookMetaData = agent.getBookData(&book)
			status.StartTime = borrowBook.StartTime
			status.EndTime = borrowBook.EndTime
			deadline := util.StringToTime(borrowBook.StartTime).Add(time.Hour * 240)
			status.Deadline = deadline.Format(util.GormTimeFormat)
			status.Fine = CalculateFine(status)
			statusArr = append(statusArr, status)
		}
	}
	return statusArr
}

func (agent *DBAgent) GetMemberReserveBooksPages(userId int) int64 {
	var count int64
	if err := agent.DB.Table("reserve").Where("user_id = ?", userId).Count(&count).Error; err != nil {
		return 0
	}
	return (count - 1) / itemsPerPage + 1
}

func (agent *DBAgent) GetMemberReserveBooks(userId int, page int) []ReserveBookStatus {
	reserveBooks := make([]ReserveBook, 0)
	statusArr := make([]ReserveBookStatus, 0)
	agent.DB.Where("user_id = ?", userId).Find(&reserveBooks).Offset((page - 1) * itemsPerPage).Limit(itemsPerPage)
	for _, reserveBook := range reserveBooks {
		book := Book{}
		if err := agent.DB.First(&book, reserveBook.BookId).Error; err == nil {
			status := ReserveBookStatus{}
			status.BookMetaData = agent.getBookData(&book)
			status.StartTime = reserveBook.StartTime
			status.EndTime = reserveBook.EndTime
			if reserveBook.EndTime == "" {
				canceledTime := util.StringToTime(reserveBook.StartTime).Add(time.Hour * time.Duration(ReserveHours))
				status.CanceledTime = canceledTime.Format(util.GormTimeFormat)
			}
			statusArr = append(statusArr, status)
		}
	}
	return statusArr
}

func (agent *Agent) BorrowBook(userId int, bookId int) StatusResult {
	result := StatusResult{}
	GetMemberFine(agent.DB, userId)
	err := agent.DB.Transaction(func(tx *gorm.DB) error {
		// 判断图书状态
		book := Book{}
		if err := tx.First(&book, bookId).Error; err != nil {
			return err
		}
		if book.State == Damaged {
			return errors.New("图书损毁")
		}
		if book.State == Lost {
			return errors.New("图书遗失")
		}
		if book.State == Borrowed {
			return errors.New("借阅失败，该图书已经被借出")
		}

		// 判断是否有罚金未缴纳
		user := User{}
		if err := tx.Find(&user, userId).Error; err != nil {
			return err
		}
		if user.Debt != 0 {
			return errors.New("请先缴纳罚金")
		}

		// 判断是否借阅达到了五本书
		var count int64
		tx.Model(&BorrowBook{}).Where("user_id = ? and end_time is null", userId).Count(&count)
		if count >= 5 {
			return errors.New("达到借阅上限")
		}

		// 判断该书是否被预约
		if book.State == Reserved {
			reserveBook := ReserveBook{}
			if err := tx.Where("book_id = ?", bookId).Last(&reserveBook).Error; err != nil {
				return err
			}
			// 是否被其他人预约
			if reserveBook.UserId != userId {
				return errors.New("借阅失败，该图书已经被预约")
			} else {
				reserveBook.EndTime = util.TimeToString(time.Now())
				if err := tx.Model(&reserveBook).Select("end_time").Updates(&reserveBook).Error; err != nil {
					return err
				}
			}
		}

		// 借阅
		newBorrowBook := BorrowBook{
			BookId:    bookId,
			UserId:    userId,
			StartTime: util.TimeToString(time.Now()),
		}
		if err := tx.Select("book_id", "user_id", "start_time").Create(&newBorrowBook).Error; err != nil {
			return err
		}
		book.State = Borrowed
		if err := tx.Model(&book).Select("state").Updates(&book).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		result.Status = BorrowFailed
		result.Msg = err.Error()
	} else {
		result.Status = BorrowOK
		result.Msg = "借阅成功"
	}
	return result
}

func (agent *DBAgent) ReturnBook(bookId int) StatusResult {
	result := StatusResult{}
	err := agent.DB.Transaction(func(tx *gorm.DB) error {
		book := Book{}
		if err := tx.First(&book, bookId).Error; err != nil {
			return err
		}
		if book.State != Borrowed {
			return errors.New("图书未被借阅")
		}
		borrowBook := BorrowBook{}
		if err := tx.Where("book_id = ?", bookId).Last(&borrowBook).Error; err != nil {
			return err
		}
		borrowBook.EndTime = util.TimeToString(time.Now())
		if err := tx.Model(&borrowBook).Select("end_time").Updates(&borrowBook).Error; err != nil {
			return err
		}
		book.State = Idle
		if err := tx.Model(&book).Select("state").Updates(&book).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		result.Status = ReturnFailed
		result.Msg = err.Error()
	} else {
		result.Status = ReturnOK
		result.Msg = "归还成功"
	}
	return result
}

func (agent *DBAgent) ReserveBook(userId int, bookId int) StatusResult {
	result := StatusResult{}
	user := User{}
	book := Book{}
	err := agent.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&user, userId).Error; err != nil {
			return err
		}
		// 判断图书状态
		if err := tx.First(&book, bookId).Error; err != nil {
			return err
		}
		if book.State == Damaged {
			return errors.New("预约失败，图书已经损毁")
		}
		if book.State == Lost {
			return errors.New("预约失败，图书已经遗失")
		}
		if book.State == Reserved {
			return errors.New("预约失败，图书已经被预约")
		}
		if book.State == Borrowed {
			return errors.New("预约失败，图书已经被借出")
		}

		newReserveBook := &ReserveBook{
			BookId:    bookId,
			UserId:    userId,
			StartTime: util.TimeToString(time.Now()),
		}
		if err := tx.Select("book_id", "user_id", "start_time").Create(&newReserveBook).Error; err != nil {
			return err
		}
		book.State = Reserved
		if err := tx.Model(&book).Select("state").Updates(&book).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		result.Status = ReserveFailed
		result.Msg = err.Error()
	} else {
		result.Status = ReserveOK
		result.Msg = "预约成功"
		title := "预约图书成功"
		endTime := time.Now().Add(time.Hour * time.Duration(ReserveHours))
		content := fmt.Sprintf("您成功预约图书《%s》，请在%s前取出图书", book.Name, util.TimeToString(endTime))
		go SendEmail(user.Email, title, content)
	}
	return result
}

func (agent *DBAgent) CancelReserveBook(userId int, bookId int) StatusResult {
	result := StatusResult{}
	user := User{}
	book := Book{}
	err := agent.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&user, userId).Error; err != nil {
			return err
		}
		if err := tx.First(&book, bookId).Error; err != nil {
			return err
		}
		if book.State != Reserved {
			return errors.New("取消预约失败，该图书未被预约")
		}
		reserveBook := ReserveBook{}
		if err := tx.Where("book_id = ?", bookId).Last(&reserveBook).Error; err != nil {
			return err
		}

		if userId != reserveBook.UserId {
			return errors.New("取消预约失败，该图书不是你预约的")
		}

		reserveBook.EndTime = util.TimeToString(time.Now())
		if err := tx.Model(&reserveBook).Select("end_time").Updates(&reserveBook).Error; err != nil {
			return err
		}
		book.State = Idle
		if err := tx.Model(&book).Select("state").Updates(&book).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		result.Status = CancelReserveFailed
		result.Msg = err.Error()
	} else {
		result.Status = CancelReserveOK
		result.Msg = "取消预约成功"
		title := "取消预约图书成功"
		content := fmt.Sprintf("您已取消预约图书《%s》", book.Name)
		go SendEmail(user.Email, title, content)
	}
	return result
}

func (agent *DBAgent) GetMemberHistoryBorrowTime(userId int) int {
	borrowBooks := make([]BorrowBook, 0)
	agent.DB.Where("user_id = ?", userId).Find(&borrowBooks)
	days := 0
	for _, borrowBook := range borrowBooks {
		startTime := util.StringToTime(borrowBook.StartTime)
		var endTime time.Time
		if borrowBook.EndTime != "" {
			endTime = util.StringToTime(borrowBook.EndTime)
		} else {
			endTime = time.Now()
		}
		days += CalculateDayDiff(startTime, endTime)
	}
	return days
}

func GetMemberFine(db *gorm.DB, userId int) int {
	user := User{}
	borrowBooks := make([]BorrowBook, 0)
	pays := make([]Pay, 0)
	_ = db.Transaction(func(tx *gorm.DB) error {
		fine := 0
		paid := 0
		tx.First(&user, userId)
		tx.Where("user_id = ?", userId).Find(&borrowBooks)
		tx.Where("user_id = ? and done = ?", userId, 1).Find(&pays)
		for _, borrowBook := range borrowBooks {
			status := BorrowBookStatus{}
			status.StartTime = borrowBook.StartTime
			status.EndTime = borrowBook.EndTime
			fine += CalculateFine(status)
		}
		for _, pay := range pays {
			paid += pay.Amount
		}
		user.Debt = fine - paid
		tx.Model(&user).Select("debt").Updates(&user)
		return nil
	})
	return user.Debt
}

func GetMemberHistoryFine(db *gorm.DB, userId int) int {
	borrowBooks := make([]BorrowBook, 0)
	fine := 0
	_ = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? and fine > -1", userId).Find(&borrowBooks).Error; err != nil {
			return err
		}
		for _, borrowBook := range borrowBooks {
			status := BorrowBookStatus{}
			status.StartTime = borrowBook.StartTime
			status.EndTime = borrowBook.EndTime
			fine += CalculateFine(status)
		}
		return nil
	})
	return fine
}

func (agent *DBAgent) GetMemberCurrentBorrowCount(userId int) int {
	var count int64
	agent.DB.Model(&BorrowBook{}).Where("user_id = ? and end_time is null", userId).Count(&count)
	return int(count)
}

func (agent *DBAgent) GetMemberCurrentReserveCount(userId int) int {
	var count int64
	agent.DB.Model(&ReserveBook{}).Where("user_id = ? and end_time is null", userId).Count(&count)
	return int(count)
}

func (agent *Agent) GetPayMemberFineURL(userId int) (urlStr string) {
	payClient := agent.PayClient
	db := agent.DB
	pay := Pay{}

	fine := GetMemberFine(agent.DB, userId)

	if fine == 0 {
		return
	}

	createNewPay := false
	err := db.Transaction(func(tx *gorm.DB) error {
		// 判断是否有未支付的
		if err := tx.Where("user_id = ?", userId).Last(&pay).Error; err == nil {
			if fine != pay.Amount && pay.Done != 1 {
				createNewPay = true
				pay.Done = -1
				if err := tx.Model(&pay).Select("done").Updates(&pay).Error; err != nil {
					return err
				}
			}
		}
		if createNewPay {
			// 创建新的pay
			user := User{}
			if err := tx.First(&user, userId).Error; err != nil {
				return err
			}
			pay.UserId = userId
			pay.Amount = user.Debt
			if err := tx.Select("user_id", "amount").Create(&pay).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return
	}
	if pay.Amount == 0 {
		return
	}
	var p = alipay.TradePagePay{}
	p.Subject = "罚金"
	p.OutTradeNo = strconv.Itoa(pay.Id)
	p.TotalAmount = strconv.Itoa(pay.Amount)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	// p.NotifyURL = "http://121.5.46.215/alipayNotify"
	u, err := payClient.TradePagePay(p)
	if err != nil {
		log.Println(err)
		return
	}
	urlBytes, err := u.MarshalBinary()
	if err != nil {
		return
	}
	log.Println("New pay: ", pay.Id, pay.Amount)
	urlStr = string(urlBytes)
	return
}

func (agent *DBAgent) GetMemberHistoryFineListPages(userId int) int64 {
	var count int64
	if err := agent.DB.Table("pay").Where("user_id = ? and done > -1", userId).Count(&count).Error; err != nil {
		return 0
	}
	return (count - 1) / itemsPerPage + 1
}

func (agent *DBAgent) GetMemberHistoryFineListByPage(userId int, page int) []FineData {
	fineDataArr := make([]FineData, 0)
	pays := make([]Pay, 0)
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("user_id = ? and done > -1", userId).Offset((page - 1) * itemsPerPage).Limit(itemsPerPage).Find(&pays)
		for _, pay := range pays {
			fineData := FineData{
				UserId:    userId,
				Fine:      pay.Amount,
				Done:      pay.Done,
			}
			fineDataArr = append(fineDataArr, fineData)
		}
		return nil
	})
	return fineDataArr
}

func (agent *DBAgent) GetMemberReturnHistoryPages(userId int) int {
	var count int64
	agent.DB.Model(&BorrowBook{}).Where("user_id = ? and end_time is not null", userId).Count(&count)
	return int((count - 1) / itemsPerPage + 1)
}

func (agent *DBAgent) GetMemberReturnHistory(userId int, page int) []BorrowBookStatus {
	borrowBooks := make([]BorrowBook, 0)
	statusArr := make([]BorrowBookStatus, 0)
	agent.DB.Where("user_id = ? and end_time is not null", userId).Find(&borrowBooks).Offset((page - 1) * itemsPerPage).Limit(itemsPerPage)
	for _, borrowBook := range borrowBooks {
		book := Book{}
		if err := agent.DB.First(&book, borrowBook.BookId).Error; err == nil {
			status := BorrowBookStatus{}
			status.BookMetaData = agent.getBookData(&book)
			status.StartTime = borrowBook.StartTime
			status.EndTime = borrowBook.EndTime
			deadline := util.StringToTime(borrowBook.StartTime).Add(time.Hour * 240)
			status.Deadline = deadline.Format(util.GormTimeFormat)
			status.Fine = CalculateFine(status)
			statusArr = append(statusArr, status)
		}
	}
	return statusArr
}
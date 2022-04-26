package services

import (
	"github.com/smartwalle/alipay/v3"
	"gorm.io/gorm"
	"lms/util"
	"math"
	"strconv"
	"time"
)

type User struct {
	UserId   int    `db:"id" gorm:"column:id;primaryKey"`
	Username string `db:"username" gorm:"column:username"`
	Password string `db:"password" gorm:"column:password"`
	Email    string `db:"email" gorm:"column:email"`
	Debt     int    `db:"debt" gorm:"column:debt"`
}

func (user User) TableName() string {
	return "user"
}

type LoginInterface func(userId int, password string) StatusResult

func (agent *DBAgent) AuthenticateUser(userId int, password string) StatusResult {
	result := StatusResult{}
	user := &User{}
	tx := agent.DB.First(&user, userId)
	if tx.Error != nil {
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
	return count / 10 + 1
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
	agent.DB.Where("user_id = ?", userId).Find(&borrowBooks).Offset((page - 1) * 10).Limit(10)
	for _, borrowBook := range borrowBooks {
		book := Book{}
		if err := agent.DB.First(&book, borrowBook.BookId).Error; err == nil {
			status := BorrowBookStatus{}
			status.Book = book
			status.StartTime = borrowBook.StartTime
			status.EndTime = borrowBook.EndTime
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
	return count / 10 + 1
}

func (agent *DBAgent) GetMemberReserveBooks(userId int, page int) []ReserveBookStatus {
	reserveBooks := make([]ReserveBook, 0)
	statusArr := make([]ReserveBookStatus, 0)
	agent.DB.Where("user_id = ?", userId).Find(&reserveBooks).Offset((page - 1) * 10).Limit(10)
	for _, reserveBook := range reserveBooks {
		book := Book{}
		if err := agent.DB.First(&book, reserveBook.BookId).Error; err == nil {
			status := ReserveBookStatus{}
			status.Book = book
			status.StartTime = reserveBook.StartTime
			status.EndTime = reserveBook.EndTime
			statusArr = append(statusArr, status)
		}
	}
	return statusArr
}

func (agent *Agent) BorrowBook(userId int, bookId int) StatusResult {
	result := StatusResult{}
	agent.GetMemberFine(userId)
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		// 判断bookId
		if err := tx.First(&Book{}, bookId).Error; err != nil {
			result.Status = BorrowFailed
			result.Msg = "借阅失败"
			return err
		}

		// 判断是否有罚金未缴纳
		user := User{}
		if err := tx.Find(&user, userId).Error; err != nil {
			return err
		}
		if user.Debt != 0 {
			result.Status = BorrowFailed
			result.Msg = "请先缴纳罚金"
			return nil
		}

		// 判断是否借阅达到了五本书
		var count int64
		tx.Model(&BorrowBook{}).Where("user_id = ? and end_time is null", userId).Count(&count)
		if count >= 5 {
			result.Status = BorrowFailed
			result.Msg = "达到借阅上限"
			return nil
		}

		// 判断该书是否被预约
		reserveBook := ReserveBook{}
		if err := tx.Where("book_id = ?", bookId).Last(&reserveBook).Error; err == nil {
			// 判断是否被预约
			if reserveBook.EndTime == "" {
				// 是否被其他人预约
				if reserveBook.UserId != userId {
					result.Status = BorrowFailed
					result.Msg = "借阅失败，该图书已经被预约"
					return nil
				} else {
					reserveBook.EndTime = util.TimeToString(time.Now())
					if err := tx.Model(&reserveBook).Select("end_time").Updates(&reserveBook).Error; err != nil {
						result.Status = BorrowFailed
						result.Msg = "借阅失败"
						return err
					}
					result.Status = BorrowOK
					result.Msg = "借阅成功"
				}
			}
		}

		// 再判断该书是否已经被借过
		borrowBook := BorrowBook{}
		if err := tx.Where("book_id = ?", bookId).Last(&borrowBook).Error; err == nil {
			if borrowBook.EndTime == "" {
				result.Status = BorrowFailed
				result.Msg = "借阅失败，该图书已经被借出"
				return nil
			}
		}

		newBorrowBook := BorrowBook{
			BookId:    bookId,
			UserId:    userId,
			StartTime: util.TimeToString(time.Now()),
		}

		// 借阅
		if err := tx.Select("book_id", "user_id", "start_time").Create(&newBorrowBook).Error; err != nil {
			result.Status = BorrowFailed
			result.Msg = "借阅失败"
			return err
		}

		result.Status = BorrowOK
		result.Msg = "借阅成功"
		return nil
	})
	return result
}

func (agent *DBAgent) ReturnBook(userId int, bookId int) StatusResult {
	result := StatusResult{}
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		borrowBook := BorrowBook{}
		if err := tx.Where("book_id = ?", bookId).Last(&borrowBook).Error; err == nil {
			if borrowBook.EndTime != "" {
				result.Status = ReturnFailed
				result.Msg = "归还失败，该图书未被借阅"
				return nil
			}

			if userId != borrowBook.UserId {
				result.Status = ReturnFailed
				result.Msg = "归还失败，该图书不是借给你的"
				return nil
			}

			borrowBook.EndTime = util.TimeToString(time.Now())
			if err := tx.Model(&borrowBook).Select("end_time").Updates(&borrowBook).Error; err != nil {
				result.Status = ReturnFailed
				result.Msg = "归还失败"
				return err
			}

			result.Status = ReturnOK
			result.Msg = "归还成功"
			return nil
		}
		result.Status = ReturnFailed
		result.Msg = "归还失败，该图书未被借阅"
		return nil
	})
	return result
}

func (agent *DBAgent) ReserveBook(userId int, bookId int) StatusResult {
	result := StatusResult{}
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		// 判断bookId
		if err := tx.First(&Book{}, bookId).Error; err != nil {
			result.Status = ReserveFailed
			result.Msg = "预约失败"
			return err
		}

		reserveBook := ReserveBook{}
		// 判断该书是否已经被预约
		if err := tx.Where("book_id = ?", bookId).Last(&reserveBook).Error; err == nil {
			if reserveBook.EndTime == "" {
				result.Status = ReserveFailed
				result.Msg = "预约失败，该图书已经被预约了"
				return nil
			}
		}

		// 判断该书是否已经被借出
		borrowBook := &BorrowBook{}
		if err := tx.Where("book_id = ?", bookId).Last(&borrowBook).Error; err == nil {
			if borrowBook.EndTime == "" {
				result.Status = ReserveFailed
				result.Msg = "预约失败，该图书已经被借出了"
				return nil
			}
		}

		newReserveBook := &ReserveBook{
			BookId: bookId,
			UserId: userId,
			StartTime: util.TimeToString(time.Now()),
		}
		if err := tx.Select("book_id", "user_id", "start_time").Create(&newReserveBook).Error; err != nil {
			result.Status = ReserveFailed
			result.Msg = "预约失败"
			return err
		}

		result.Status = ReserveOK
		result.Msg = "预约成功"
		return nil
	})
	return result
}

func (agent *DBAgent) CancelReserveBook(userId int, bookId int) StatusResult {
	result := StatusResult{}
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
		reserveBook := ReserveBook{}
		// 先判断该书是否已经被预约
		if err := tx.Where("book_id = ?", bookId).Last(&reserveBook).Error; err == nil {
			if reserveBook.EndTime != "" {
				result.Status = CancelReserveFailed
				result.Msg = "取消预约失败，该图书未被预约"
				return nil
			}

			if userId != reserveBook.UserId {
				result.Status = CancelReserveFailed
				result.Msg = "取消预约失败，该图书不是被你预约的"
				return nil
			}

			reserveBook.EndTime = util.TimeToString(time.Now())
			if err := tx.Model(&reserveBook).Select("end_time").Updates(&reserveBook).Error; err != nil {
				result.Status = CancelReserveFailed
				result.Msg = "取消预约失败"
				return err
			}

			result.Status = CancelReserveOK
			result.Msg = "取消预约成功"
			return nil
		} else {
			result.Status = CancelReserveFailed
			result.Msg = "取消预约失败，该图书未被预约"
			return nil
		}
	})
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

func (agent *Agent) GetMemberFine(userId int) int {
	user := User{}
	borrowBooks := make([]BorrowBook, 0)
	pays := make([]Pay, 0)
	_ = agent.DB.Transaction(func(tx *gorm.DB) error {
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

func (agent *Agent) GetPayMemberFineURL(userId int) (urlStr string) {
	payClient := agent.PayClient
	db := agent.DB
	pay := Pay{}

	agent.GetMemberFine(userId)

	hasOldPay := false
	err := db.Transaction(func(tx *gorm.DB) error {
		// 判断是否有未支付的
		if err := tx.Where("user_id = ?", userId).Last(&pay).Error; err == nil {
			if pay.Done == 0 {
				hasOldPay = true
			}
		}
		if !hasOldPay {
			// 创建新的pay
			user := User{}
			tx.First(&user, userId)
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
	var p = alipay.TradePagePay{}
	p.Subject = "罚金"
	p.OutTradeNo = strconv.Itoa(pay.Id)
	p.TotalAmount = strconv.Itoa(pay.Amount)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	u, err := payClient.TradePagePay(p)
	if err != nil {
		return
	}
	urlBytes, err := u.MarshalBinary()
	if err != nil {
		return
	}
	urlStr = string(urlBytes)
	return
}

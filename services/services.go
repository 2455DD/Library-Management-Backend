package services

import (
	"database/sql"
	"fmt"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"image/png"
	"os"
	"path/filepath"
	"time"
)

type DBAgent struct {
	DB *sql.DB
}

type User struct {
	UserID   int    `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
}

type Book struct {
	Id       int    `db:"id"`
	Name     string `db:"name"`
	Author   string `db:"author"`
	Isbn     string `db:"isbn"`
	Address  string `db:"address"`
	Language string `db:"language"`
	Count    int    `db:"count"`
	Location string `db:"location"`
}

type StatusResult struct {
	Code   int
	Msg    string
	Status StatusCode
}

type StatusCode int

const (
	OK StatusCode = iota
	Error
	UsernameOrPasswordError
	UserLoginOK
	AdminLoginOK
	UsernameExist
	RegisterError
	RegisterOK
	BorrowNotEnough
	BorrowFailed
	BorrowOK
	ReturnFailed
	ReturnOK
	UpdateFailed
	UpdateOK
	DeleteFailed
	DeleteOK
	UserBarcodeFailed
	UserBarcodeOK
)

var (
	mediaPath string
)

func (agent DBAgent) AuthenticateAdmin(username string, password string) (*StatusResult, int) {
	result := &StatusResult{}
	user := new(User)
	user.Username = username
	row := agent.DB.QueryRow(fmt.Sprintf("select * from admin where username='%v'", user.Username))
	err := row.Scan(&user.UserID, &user.Username, &user.Password)
	if err != nil {
		result.Status = UsernameOrPasswordError
		result.Msg = "不存在此用户"
		return result, -1
	}
	if password != user.Password {
		result.Status = UsernameOrPasswordError
		result.Msg = "密码错误"
		return result, -1
	}
	result.Status = AdminLoginOK
	result.Msg = "用户登录成功"
	return result, user.UserID
}

func (agent DBAgent) AuthenticateUser(username string, password string) (*StatusResult, int) {
	result := &StatusResult{}
	user := new(User)
	user.Username = username
	row := agent.DB.QueryRow(fmt.Sprintf("select * from user where username='%v'", user.Username))
	err := row.Scan(&user.UserID, &user.Username, &user.Password)
	if err != nil {
		result.Status = UsernameOrPasswordError
		result.Msg = "不存在此用户"
		return result, -1
	}
	if password != user.Password {
		result.Status = UsernameOrPasswordError
		result.Msg = "密码错误"
		return result, -1
	}
	result.Status = UserLoginOK
	result.Msg = "用户登录成功"
	return result, user.UserID
}

func (agent DBAgent) RegisterUser(username string, password string) (*StatusResult, string) {
	result := &StatusResult{}
	if agent.HasUser(username) {
		result.Status = UsernameExist
		result.Msg = "用户名已经存在"
		return result, ""
	}
	command := fmt.Sprintf("insert INTO user(username, password) values('%v','%v')", username, password)
	_, err := agent.DB.Exec(command)
	if err != nil {
		result.Status = RegisterError
		result.Msg = "注册失败"
		fmt.Println(err.Error())
		return result, ""
	}

	path, getBarcodeResult := agent.GetUserBarcodePath(username)
	if (path == "fail") || (getBarcodeResult.Status == UserBarcodeFailed) {
		return &StatusResult{
			Msg:    "获取用户条形码存储路径失败",
			Status: UserBarcodeFailed,
		}, ""
	}

	result.Status = RegisterOK
	result.Msg = "注册成功"

	return result, path
}

func (agent DBAgent) HasUser(username string) bool {
	user := new(User)
	row := agent.DB.QueryRow(fmt.Sprintf("select username from user where username='%v'", username))
	err := row.Scan(&user.Username)
	if err != nil {
		row = agent.DB.QueryRow(fmt.Sprintf("select username from admin where username='%v'", username))
		err := row.Scan(&user.Username)
		if err != nil {
			return false
		}
	}
	return true
}

func (agent DBAgent) GetUserBarcodePath(username string) (string, *StatusResult) {
	//检查用户的username是否在数据库中
	if !agent.HasUser(username) {
		return "fail", &StatusResult{
			Msg:    "数据库中不存在该username",
			Status: UserBarcodeFailed,
		}
	}

	var path string
	row := agent.DB.QueryRow(fmt.Sprintf("SELECT barcode_path from user_barcode WHERE id='%v';", username))
	err := row.Scan(&path)
	if err != nil {
		return "fail", &StatusResult{
			Msg:    "用户条形码不存在",
			Status: UserBarcodeFailed,
		}
	}
	return path, &StatusResult{
		Msg:    "成功获取",
		Status: UserBarcodeOK,
	}

}

func (agent DBAgent) StoreUserBarcodePath(username string) *StatusResult {
	//检查用户的username是否在数据库中
	if !agent.HasUser(username) {
		return &StatusResult{
			Msg:    "数据库中不存在该username",
			Status: UserBarcodeFailed,
		}
	}

	path, generateResult := agent.GenerateUserBarcode(username)
	if (path == "fail") || (generateResult.Status == UserBarcodeFailed) {
		return &StatusResult{
			Msg:    "生成用户条形码失败",
			Status: UserBarcodeFailed,
		}
	}

	return &StatusResult{
		Msg:    "生成用户条形码失败",
		Status: UserBarcodeFailed,
	}

}

func (agent DBAgent) GenerateUserBarcode(username string) (string, *StatusResult) {
	result := &StatusResult{}

	//创建一个code128编码的 BarcodeIntCS
	cs, err := code128.Encode(username)
	if err != nil {
		result.Status = UserBarcodeFailed
		result.Msg = "用户条形码编码失败"
		return "fail", result
	}

	//创建一个要输出数据的文件
	path := filepath.Join(mediaPath, fmt.Sprintf("%v.png", username))
	file, err := os.Create(path)
	if err != nil {
		result.Status = UserBarcodeFailed
		result.Msg = "生成用户二维码PNG文件失败"
		return "fail", result
	}

	defer file.Close()

	// 设置图片像素大小
	qrCode, _ := barcode.Scale(cs, 350, 70)
	// 将code128的条形码编码为png图片
	png.Encode(file, qrCode)

	result.Status = UserBarcodeOK
	result.Msg = "用户二维码生成成功"

	return path, result
}

func (agent DBAgent) GetBookNum() int {
	command := "SELECT COUNT(*) FROM book"
	row, _ := agent.DB.Query(command)
	count := 0
	for row.Next() {
		count += 1
	}
	return count
}

func (agent DBAgent) GetBorrowTime(bookId int) int {
	command := fmt.Sprintf("select a.createtime from borrow a where a.book_id=%v;", bookId)
	row, err := agent.DB.Query(command)
	if err != nil {
		return 0
	}
	var subTime time.Duration = 0
	var creatTime time.Time
	for row.Next() {
		err = row.Scan(&creatTime)
		currentTime := time.Now()
		if err != nil {
			fmt.Println(err.Error())
			return 0
		}
		subTime = currentTime.Sub(creatTime)
	}
	return int(subTime.Hours() / 24)
}

func (agent DBAgent) GetBooksByPage(page int) []Book {
	// 1页10条
	command := fmt.Sprintf("SELECT * FROM book limit %v,10;", page/10)
	row, err := agent.DB.Query(command)
	books := make([]Book, 0, 10)
	if err != nil {
		fmt.Println(err.Error())
		return books
	}
	for row.Next() {
		book := Book{}
		err := row.Scan(&book.Id, &book.Name, &book.Author, &book.Isbn, &book.Address, &book.Language, &book.Count)
		if err != nil {
			return books
		}
		books = append(books, book)
	}
	return books
}

func (agent DBAgent) GetUserBooksByPage(userID int, page int) []Book {
	// 1页10条
	command := fmt.Sprintf("select a.* from book a inner join borrow b on a.id=b.book_id and b.user_id=%v limit %v,10;", userID, page/10)
	row, err := agent.DB.Query(command)
	books := make([]Book, 0, 10)
	if err != nil {
		fmt.Println(err.Error())
		return books
	}
	for row.Next() {
		book := Book{}
		err := row.Scan(&book.Id, &book.Name, &book.Author, &book.Isbn, &book.Address, &book.Language, &book.Count)
		if err != nil {
			return books
		}
		books = append(books, book)
	}
	return books
}

func (agent DBAgent) BorrowBook(userID int, bookID int) *StatusResult {
	result := &StatusResult{}
	var bookCount int
	row := agent.DB.QueryRow(fmt.Sprintf("select count from book where id=%v", bookID))
	err := row.Scan(&bookCount)
	if err != nil {
		fmt.Println(err.Error())
		result.Status = BorrowFailed
		result.Msg = "借阅失败"
		return result
	}
	if bookCount == 0 {
		result.Status = BorrowNotEnough
		result.Msg = "借阅失败，图书数量不足"
		return result
	}

	tx, _ := agent.DB.Begin()

	ret1, _ := tx.Exec(fmt.Sprintf("insert into borrow(user_id, book_id) values(%v,%v)", userID, bookID))
	insNums, _ := ret1.RowsAffected()

	ret2, _ := tx.Exec(fmt.Sprintf("UPDATE book set count=count-1 where id=%v", bookID))
	updNums, _ := ret2.RowsAffected()

	if insNums > 0 && updNums > 0 {
		_ = tx.Commit()

		result.Status = BorrowOK
		result.Msg = "借阅成功"
		return result
	} else {
		_ = tx.Rollback()

		result.Status = BorrowFailed
		result.Msg = "借阅失败"
		return result
	}

}

func (agent DBAgent) ReturnBook(userID int, bookID int) *StatusResult {
	result := &StatusResult{}
	var borrowID int
	row := agent.DB.QueryRow(fmt.Sprintf("select id from borrow where user_id=%v and book_id=%v", userID, bookID))
	err := row.Scan(&borrowID)
	if err != nil {
		fmt.Println(err.Error())
		result.Status = ReturnFailed
		result.Msg = "归还失败，你没有借阅该书籍"
		return result
	}

	tx, _ := agent.DB.Begin()

	ret1, _ := tx.Exec(fmt.Sprintf("delete from borrow where user_id=%v and book_id=%v limit 1", userID, bookID))
	delNums, _ := ret1.RowsAffected()

	ret2, _ := tx.Exec(fmt.Sprintf("UPDATE book set count=count+1 where id=%v", bookID))
	updNums, _ := ret2.RowsAffected()

	if delNums > 0 && updNums > 0 {
		_ = tx.Commit()

		result.Status = ReturnOK
		result.Msg = "归还成功"
		return result
	} else {
		_ = tx.Rollback()

		result.Status = ReturnFailed
		result.Msg = "归还失败"
		return result
	}
}

func (agent DBAgent) UpdateBookStatus(newBook *Book) *StatusResult {
	result := &StatusResult{}
	book := new(Book)
	var nums int64
	row := agent.DB.QueryRow(fmt.Sprintf("select * from borrow where id=%v", book.Id))
	err := row.Scan(&book.Id, &book.Name, &book.Author, &book.Isbn, &book.Address, &book.Language, &book.Count)
	if err != nil {
		// add book
		command := fmt.Sprintf(
			"insert into book(name, author, isbn, address, language, count) values('%v','%v','%v','%v','%v',%v)",
			newBook.Name, newBook.Author, newBook.Isbn, newBook.Address, newBook.Language, newBook.Count)
		ret, _ := agent.DB.Exec(command)
		nums, _ = ret.RowsAffected()
	} else {
		//update book
		command := fmt.Sprintf(
			"update book set name='%v', author='%v', isbn='%v', address='%v', language='%v', count=%v where id=%v",
			newBook.Name, newBook.Author, newBook.Isbn, newBook.Address, newBook.Language, newBook.Count, newBook.Id)
		ret, _ := agent.DB.Exec(command)
		nums, _ = ret.RowsAffected()
	}
	if nums > 0 {
		result.Status = UpdateOK
		result.Msg = "更新成功"
		return result
	} else {
		result.Status = UpdateFailed
		result.Msg = "更新失败"
		return result
	}
}

func (agent DBAgent) AddBook(newBook *Book) (result *StatusResult) {
	if newBook == nil {
		return &StatusResult{
			Msg:    "parameter is a nil pointer",
			Status: UpdateFailed,
		}
	}
	result = new(StatusResult)
	row := agent.DB.QueryRow(fmt.Sprintf("SELECT * FROM book WHERE book.isbn=%v", newBook.Isbn))
	err := row.Scan()
	// If Exists
	if err == nil {
		result.Status = UpdateFailed
		result.Msg = "加入失败, 数据库内已有该isbn号的书"
		return result
	}

	err = nil
	transaction, err := agent.DB.Begin()
	if err != nil {
		result.Status = UpdateFailed
		result.Msg = "加入失败, 数据库语句出错，信息如下\n" + fmt.Sprintln(err.Error())
		return result
	}
	var ret sql.Result
	for i := 0; i < newBook.Count; i++ {
		ret, err = transaction.Exec(fmt.Sprintf("INSERT INTO "+
			"book(name, author, isbn , language, count, location) "+
			"VALUES('%v','%v','%v','%v','%v','%v')",
			newBook.Name, newBook.Author, newBook.Isbn, newBook.Language, newBook.Count, newBook.Location))
		if err != nil {
			result.Status = UpdateFailed
			result.Msg = "加入失败, 数据库语句出错，信息如下\n" + fmt.Sprintln(err.Error())
			err = transaction.Rollback()
			if err != nil {
				result.Msg += "回滚失败!信息如下\n" + fmt.Sprintln(err)
			}
			return result
		}
		if num, _ := ret.RowsAffected(); num <= 0 {
			result.Status = UpdateFailed
			result.Msg = "加入失败, Row affected为0"
			err = transaction.Rollback()
			if err != nil {
				result.Msg += "回滚失败!信息如下\n" + fmt.Sprintln(err)
			}
			return result
		}
	}
	transaction.Commit()
	result.Status = UpdateOK
	result.Msg = "加入成功"
	return result
}

func (agent DBAgent) DeleteBook(bookID int) *StatusResult {
	result := new(StatusResult)

	tx, _ := agent.DB.Begin()

	delRet, _ := tx.Exec(fmt.Sprintf("delete from book where id=%v", bookID))
	delNums, _ := delRet.RowsAffected()

	clearRet, _ := tx.Exec(fmt.Sprintf("delete from borrow where book_id=%v", bookID))
	_, err := clearRet.RowsAffected()

	if delNums > 0 && err == nil {
		_ = tx.Commit()

		result.Status = DeleteOK
		result.Msg = "删除成功"
		return result
	} else {
		_ = tx.Rollback()

		result.Status = DeleteFailed
		result.Msg = "删除失败"
		return result
	}
}

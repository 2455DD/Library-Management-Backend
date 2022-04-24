package services

import (
	"database/sql"
	"fmt"
	"time"
)

// hasFine 检查是否有罚金未交
// para: userID-用户ID
// return: bool-返回true=有未交的罚金;返回false=无未交的罚金
func (agent DBAgent) hasFine(userID int) bool {
	command := fmt.Sprintf("select a.createtime from borrow a where a.user_id=%v;", userID)
	row, err := agent.DB.Query(command)
	if err != nil {
		return true
	}

	var subTime time.Duration = 0
	var creatTime time.Time
	for row.Next() {
		err = row.Scan(&creatTime)
		//ctime, _ := time.Parse("2006-01-02 15:04:05", string(creatTime))
		currentTime := time.Now()
		if err != nil {
			fmt.Println(err.Error())
			return true
		}
		subTime = currentTime.Sub(creatTime)
		if (subTime.Hours() / 24) > 10 {
			return true
		}
	}

	return false
}

// RenewBook 续借图书
// para: borrowID-借书订单ID, userID-用户ID, bookID-书籍ID
// return: *StatusResult-操作结果。若成功, StatusResult.Status=RenewOK; 否则StatusResult.Status=RenewFail
func (agent DBAgent) RenewBook(borrowID int, userID int, bookID int) *StatusResult {
	if agent.hasFine(userID) {
		return &StatusResult{
			Msg:    "有未支付的罚金",
			Status: RenewFailed,
		}
	}

	var exist int
	command := fmt.Sprintf("select exists(select * from borrow where id=%v and user_id=%v and book_id=%v);", borrowID, userID, bookID)
	row := agent.DB.QueryRow(command)
	if temperr := row.Scan(&exist); temperr == nil && exist != 0 {
		createTime := time.Now().Format("2006-01-02 15:04:05")

		var tx *sql.Tx
		tx = new(sql.Tx)
		tx, _ = agent.DB.Begin()
		ret2, er := tx.Exec(fmt.Sprintf("UPDATE borrow set createtime='%v', endtime=NULL, state='4' where id='%v'", createTime, borrowID))
		if er != nil {

		}
		updNums, _ := ret2.RowsAffected()
		if updNums > 0 {
			_ = tx.Commit()
			return &StatusResult{
				Msg:    "续借成功",
				Status: RenewOK,
			}
		} else {
			_ = tx.Rollback()
			return &StatusResult{
				Msg:    "续借失败",
				Status: RenewFailed,
			}
		}
	}

	return &StatusResult{
		Msg:    "续借失败",
		Status: RenewFailed,
	}

}

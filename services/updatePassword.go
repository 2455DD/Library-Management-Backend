package services

import (
	"fmt"
	"strconv"
)

func (agent DBAgent) UpdatePassword(oldpsw string, newpsw string, userid string) *StatusResult {
	useridInt, _ := strconv.Atoi(userid)
	if !agent.HasUser(useridInt) {
		return &StatusResult{
			Msg:    "用户不存在",
			Status: UpdatePasswordFailed,
		}
	}

	var exist int
	command := fmt.Sprintf("select exists(select * from user where id='%v' and password=MD5('%v'));", useridInt, oldpsw)
	row := agent.DB.QueryRow(command)
	if temperr := row.Scan(&exist); temperr == nil && exist != 0 {

		tx, _ := agent.DB.Begin()
		ret2, _ := tx.Exec(fmt.Sprintf("UPDATE user set password=MD5('%v') where id='%v'", newpsw, useridInt))
		updNums, _ := ret2.RowsAffected()
		if updNums > 0 {
			_ = tx.Commit()
			return &StatusResult{
				Msg:    "修改成功",
				Status: UpdatePasswordOK,
			}
		} else {
			_ = tx.Rollback()
			return &StatusResult{
				Msg:    "修改失败",
				Status: UpdatePasswordFailed,
			}
		}
	}

	return &StatusResult{
		Msg:    "修改失败",
		Status: UpdatePasswordFailed,
	}
}

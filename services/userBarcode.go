package services

import (
	"fmt"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
)

// GetUserBarcodePath 获取用户条形码存储路径
// para: userid-用户ID
// return: string-barcode存储路径,  *StatusResult-操作结果
//         获取用户二维码路径失败时, string="fail", StatusResult.Status=UserBarcodeFailed
func (agent DBAgent) GetUserBarcodePath(userid int) (string, *StatusResult) {
	//检查用户的userid是否在数据库中
	if !agent.HasUser(userid) {
		return "fail", &StatusResult{
			Msg:    "数据库中不存在该用户ID",
			Status: UserBarcodeFailed,
		}
	}

	var path string
	result := &StatusResult{}
	row := agent.DB.QueryRow(fmt.Sprintf("SELECT barcode_path from user_barcode WHERE id='%v';", userid))
	err := row.Scan(&path)
	//若没有二维码，则生成一个
	if err != nil {
		path, result = agent.StoreUserBarcodePath(userid)
		if result.Status != UserBarcodeFailed {
			return path, &StatusResult{
				Msg:    "成功获取",
				Status: UserBarcodeOK,
			}
		}
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

// handleSQL 处理sql语句,参考EscapeForMysql
// para: path-待处理的部分sql语句（条形码存储路径）
// return: string-处理过转义字符的sql语句
func handleSQL(path string) string {
	result := make([]byte, 0, 2*len(path))
	var temp byte

	for i := 0; i < len(path); i++ {
		c := path[i]
		temp = 0

		switch c {
		case 0:
			temp = '0'
			break
		case '\n':
			temp = 'n'
			break
		case '\r':
			temp = 'r'
			break
		case '\\':
			temp = '\\'
			break
		case '\'':
			temp = '\''
			break
		case '"':
			temp = '"'
			break
		case '\032':
			temp = 'Z'
		}

		if temp != 0 {
			result = append(result, '\\', temp)
		} else {
			result = append(result, c)
		}
	}

	return string(result)
}

// StoreUserBarcodePath 生成（调用GenerateUserBarcode）并存储用户条形码
// para: userid-用户ID
// return: *StatusResult-操作结果
func (agent DBAgent) StoreUserBarcodePath(userid int) (string, *StatusResult) {
	//检查用户的userid是否在数据库中
	if !agent.HasUser(userid) {
		return "", &StatusResult{
			Msg:    "数据库中不存在该用户ID",
			Status: UserBarcodeFailed,
		}
	}

	path, generateResult := agent.GenerateUserBarcode(userid)
	if (path == "fail") || (generateResult.Status == UserBarcodeFailed) {
		return "", &StatusResult{
			Msg:    "生成用户条形码失败",
			Status: UserBarcodeFailed,
		}
	}

	preparedPath := handleSQL(path)

	row := agent.DB.QueryRow(fmt.Sprintf("SELECT EXISTS(SELECT * from user_barcode WHERE id='%v');", userid))
	var exist int
	if temperr := row.Scan(&exist); temperr == nil && exist != 0 {
		if pathh, res := agent.GetUserBarcodePath(userid); res.Status == UserBarcodeOK && pathh == path {
			return path, &StatusResult{
				Msg:    "数据库中已经存在该用户条形码",
				Status: UserBarcodeOK,
			}
		} else {
			result, sqlerr := agent.DB.Exec(fmt.Sprintf(`UPDATE user_barcode
			SET barcode_path = '%v'
			WHERE id='%v';`,
				preparedPath, userid))
			if sqlerr != nil {
				return "", &StatusResult{
					Msg:    "SQL存储失败: " + sqlerr.Error(),
					Status: UserBarcodeFailed,
				}
			}
			if noOfRow, temperr := result.RowsAffected(); temperr != nil || noOfRow <= 0 {
				return "", &StatusResult{
					Msg:    "SQL存储失败: " + temperr.Error(),
					Status: UserBarcodeFailed,
				}
			}
			return path, &StatusResult{
				Msg:    "用户条形码存储成功",
				Status: UserBarcodeOK,
			}
		}
	}
	result, sqlerr := agent.DB.Exec(fmt.Sprintf(`INSERT INTO user_barcode(id,barcode_path) 
			VALUES ('%v','%v')`, userid, preparedPath))
	if sqlerr != nil {
		return "", &StatusResult{
			Msg:    "SQL存储失败: " + sqlerr.Error(),
			Status: UserBarcodeFailed,
		}
	}
	if noOfRow, temperr := result.RowsAffected(); temperr != nil || noOfRow <= 0 {
		return "", &StatusResult{
			Msg:    "SQL存储失败: " + temperr.Error(),
			Status: UserBarcodeFailed,
		}
	}
	return path, &StatusResult{
		Msg:    "用户条形码存储成功",
		Status: UserBarcodeOK,
	}

}

// GenerateUserBarcode 生成用户条形码
// para: userid-用户ID
// return: *StatusResult-操作结果
func (agent DBAgent) GenerateUserBarcode(userid int) (string, *StatusResult) {
	result := &StatusResult{}

	//创建一个code128编码的 BarcodeIntCS
	useridStr := strconv.Itoa(userid)
	cs, err := code128.Encode(useridStr)
	if err != nil {
		result.Status = UserBarcodeFailed
		result.Msg = "用户条形码编码失败"
		return "fail", result
	}

	//创建一个要输出数据的文件
	path := filepath.Join(mediaPath, fmt.Sprintf("%v.png", userid))
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

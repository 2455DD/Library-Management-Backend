package services

import (
	"database/sql"
	"fmt"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

func subtitleBarcode(bc barcode.Barcode) image.Image {
	fontFace := basicfont.Face7x13
	fontColor := color.RGBA{A: 255}
	margin := 5 // Space between barcode and text

	// Get the bounds of the string
	bounds, _ := font.BoundString(fontFace, bc.Content())

	widthTxt := int((bounds.Max.X - bounds.Min.X) / 64)
	heightTxt := int((bounds.Max.Y - bounds.Min.Y) / 64)

	// calc width and height
	width := widthTxt
	if bc.Bounds().Dx() > width {
		width = bc.Bounds().Dx()
	}
	height := heightTxt + bc.Bounds().Dy() + margin

	// create result img
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// draw the barcode
	draw.Draw(img, image.Rect(0, 0, bc.Bounds().Dx(), bc.Bounds().Dy()), bc, bc.Bounds().Min, draw.Over)

	// TextPt
	offsetY := bc.Bounds().Dy() + margin - int(bounds.Min.Y/64)
	offsetX := (width - widthTxt) / 2

	point := fixed.Point26_6{
		X: fixed.Int26_6(offsetX * 64),
		Y: fixed.Int26_6(offsetY * 64),
	}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(fontColor),
		Face: fontFace,
		Dot:  point,
	}
	d.DrawString(bc.Content())
	return img
}

func (agent DBAgent) HasBook(isbn string, idOptional ...int) bool {
	var (
		err   error
		value int
	)
	if len(idOptional) > 0 {
		id := idOptional[0]
		row := agent.DB.QueryRow(fmt.Sprintf("SELECT EXISTS(SELECT * from book WHERE id='%v' AND isbn='%v');", id, isbn))
		err = row.Scan(&value)
	} else {
		row := agent.DB.QueryRow(fmt.Sprintf("SELECT EXISTS(SELECT * from book WHERE isbn='%v');", isbn))
		err = row.Scan(&value)
	}
	if err != nil || value == 0 {
		return false
	}
	return true
}

func (agent *DBAgent) HasBookBarCode(id int, isbn string) *StatusResult {
	var (
		err   error
		value int
	)
	row := agent.DB.QueryRow(fmt.Sprintf("SELECT EXISTS(SELECT * from book WHERE id='%v' AND isbn='%v');",
		id, isbn))
	err = row.Scan(&value)
	if err != nil || value == 0 {
		return &StatusResult{
			Code:   0,
			Msg:    "??????????????????",
			Status: BookBarcodeFailed,
		}
	}
	return &StatusResult{
		Code:   0,
		Msg:    "???????????????",
		Status: BookBarcodeOK,
	}
}

func EscapeForSQL(sql string) string {
	dest := make([]byte, 0, 2*len(sql))
	var escape byte
	for i := 0; i < len(sql); i++ {
		c := sql[i]

		escape = 0

		switch c {
		case 0: /* Must be escaped for 'mysql' */
			escape = '0'
			break
		case '\n': /* Must be escaped for logs */
			escape = 'n'
			break
		case '\r':
			escape = 'r'
			break
		case '\\':
			escape = '\\'
			break
		case '\'':
			escape = '\''
			break
		case '"': /* Better safe than sorry */
			escape = '"'
			break
		case '\032': //?????????26,?????????32,????????????1a, /* This gives problems on Win32 */
			escape = 'Z'
		}

		if escape != 0 {
			dest = append(dest, '\\', escape)
		} else {
			dest = append(dest, c)
		}
	}

	return string(dest)
}

func (agent *DBAgent) AddBookBarcode(id int, isbn string) *StatusResult {
	// ???????????? id??????isbn?????????????????????
	if !agent.HasBook(isbn, id) {
		return &StatusResult{
			Code:   0,
			Msg:    "??????????????????????????????",
			Status: BookBarcodeFailed,
		}
	}

	codingMsg := fmt.Sprintf("%v-%v", isbn, id)
	savePath := filepath.Join(MediaPath, "bookBarcode", fmt.Sprintf("%v.png", codingMsg))

	var code barcode.Barcode
	var err error
	code, err = code128.Encode(codingMsg)
	if err != nil {
		return &StatusResult{
			Code:   1,
			Msg:    "????????????",
			Status: BookBarcodeFailed,
		}
	}
	code, err = barcode.Scale(code, 500, 100)
	img := subtitleBarcode(code)
	var pngFile *os.File

	pngFile, err = os.Create(savePath)
	if err != nil {
		return &StatusResult{
			Code:   2,
			Msg:    "??????????????????:" + err.Error(),
			Status: BookBarcodeFailed,
		}
	}
	defer func() {
		nerr := pngFile.Close()
		if nerr != nil {
			log.Println("FileSystem Error! " + nerr.Error())
		}
	}()

	err = png.Encode(pngFile, img)
	if err != nil {
		return &StatusResult{
			Code:   3,
			Msg:    "png????????????:" + err.Error(),
			Status: BookBarcodeFailed,
		}
	}

	if agent.HasBookBarCode(id, isbn).Status != BookBarcodeOK {
		// ????????????????????????,??????????????????,??????????????????
		var readPath string
		qrow := agent.DB.QueryRow(fmt.Sprintf("SELECT barcode_path from book_barcode WHERE id='%v' AND isbn='%v';", id, isbn))
		qerr := qrow.Scan(&readPath)

		if qerr != nil {
			return &StatusResult{
				Code:   4,
				Msg:    "SQL????????????: " + qerr.Error(),
				Status: BookBarcodeFailed,
			}
		}
		if readPath != savePath {
			var result sql.Result
			result, qerr = agent.DB.Exec(fmt.Sprintf(`UPDATE book_barcode
			SET barcode_path = '%v'
			WHERE id='%v' AND isbn='%v';`,
				EscapeForSQL(savePath), id, isbn))
			if qerr != nil {
				return &StatusResult{
					Code:   4,
					Msg:    "SQL????????????: " + qerr.Error(),
					Status: BookBarcodeFailed,
				}
			}
			if noOfRow, temperr := result.RowsAffected(); temperr != nil || noOfRow <= 0 {
				if temperr == nil {
					return &StatusResult{
						Msg:    "SQL????????????: noOfRow<=0",
						Status: BookBarcodeFailed,
					}
				} else {
					return &StatusResult{
						Msg:    "SQL????????????: " + temperr.Error(),
						Status: BookBarcodeFailed,
					}
				}
			}
			return &StatusResult{
				Code:   0,
				Msg:    "????????????",
				Status: BookBarcodeOK,
			}
		} else {
			return &StatusResult{
				Code:   1,
				Msg:    "????????????????????????????????????????????????",
				Status: BookBarcodeOK,
			}
		}
	}
	// ??????????????????,??????
	result, sqlerr := agent.DB.Exec(fmt.Sprintf(`INSERT INTO book_barcode(id,isbn,barcode_path) 
			VALUES ('%v','%v','%v')`,
		id, isbn, EscapeForSQL(savePath)))
	if sqlerr != nil {
		return &StatusResult{
			Msg:    "SQL????????????: " + sqlerr.Error(),
			Status: BookBarcodeFailed,
		}
	}
	if noOfRow, temperr := result.RowsAffected(); temperr != nil || noOfRow <= 0 {
		if temperr == nil {
			return &StatusResult{
				Msg:    "SQL????????????: noOfRow<=0",
				Status: BookBarcodeFailed,
			}
		} else {
			return &StatusResult{
				Msg:    "SQL????????????: " + temperr.Error(),
				Status: BookBarcodeFailed,
			}
		}

	}
	return &StatusResult{
		Msg:    "????????????",
		Status: BookBarcodeOK,
	}
}

func (agent *DBAgent) GetBookBarcodePath(id int, isbn string) (*StatusResult, string) {
	if !agent.HasBook(isbn, id) {
		return &StatusResult{
			Msg:    "Book is not found amid database",
			Status: BookBarcodeFailed,
		}, ""
	}
	var barcode_path string
	//?????????????????????????????????QueryRow???
	row := agent.DB.QueryRow(fmt.Sprintf("SELECT barcode_path from book_barcode WHERE id='%v' AND isbn='%v';", id, isbn))
	err := row.Scan(&barcode_path)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Barcode of given book is not found ,attempting to add...")
			result := agent.AddBookBarcode(id, isbn)
			log.Println("[Testing]" + result.Msg)
			if result.Status == BookBarcodeFailed {
				return &StatusResult{
					Msg:    "Tried adding bookBarcode But Failed, messages are shown below:\n" + result.Msg,
					Status: BookBarcodeFailed,
				}, ""
			}
			return agent.GetBookBarcodePath(id, isbn)
		} else {
			return &StatusResult{
				Msg:    "SQL error " + err.Error(),
				Status: BookBarcodeFailed,
			}, ""
		}
	}
	return &StatusResult{
		Msg:    "Success",
		Status: BookBarcodeOK,
	}, barcode_path

}

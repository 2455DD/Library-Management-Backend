package util

import (
	"bytes"
	"encoding/json"
)

func JsonEncode(v interface{}) []byte {
	bf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(bf)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(v)
	return bf.Bytes()
}

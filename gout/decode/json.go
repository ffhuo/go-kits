package decode

import (
	"encoding/json"
	"io"
)

type JSONDecode struct {
	obj interface{}
}

func NewJSONDecode(obj interface{}) Decoder {
	if obj == nil {
		return nil
	}
	return &JSONDecode{obj: obj}
}

func (j *JSONDecode) Decode(r io.Reader) error {
	decode := json.NewDecoder(r)
	return decode.Decode(j.obj)
}

func (j *JSONDecode) Value() interface{} {
	return j.obj
}

// JSON json decoder
func JSON(r io.Reader, obj interface{}) error {
	decode := json.NewDecoder(r)
	return decode.Decode(obj)
}

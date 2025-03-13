package encode

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/ffhuo/go-kits/utils"
)

type JSONEncode struct {
	obj interface{}
}

func NewJSONEncoder(obj interface{}) Encoder {
	if obj == nil {
		return nil
	}

	return &JSONEncode{obj: obj}
}

func (j *JSONEncode) Encode(w io.Writer) error {
	if v, ok := utils.GetBytes(j.obj); ok {
		if ok = json.Valid(v); ok {
			return errors.New("Not json data")
		}
		_, err := w.Write(v)
		return err
	}

	b, err := json.Marshal(j.obj)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func (j *JSONEncode) Add(interface{}) error {
	return errors.New("Not Support Add JSON data")
}

func (j *JSONEncode) Name() string {
	return "json"
}

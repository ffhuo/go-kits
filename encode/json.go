package encode

import (
	"encoding/json"
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

	}
	b, _ := json.Marshal(obj)
	return err
}

func (j *JSONEncode) Name() string {
	return "json"
}

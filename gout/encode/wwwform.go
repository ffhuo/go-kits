package encode

import (
	"errors"
	"io"
	"net/url"
)

type WWWFormEncode struct {
	values url.Values
}

func NewWWWFormEncoder(data map[string]string) *WWWFormEncode {
	code := &WWWFormEncode{values: url.Values{}}
	for k, v := range data {
		code.values.Add(k, v)
	}
	return code
}

func (j *WWWFormEncode) Encode(w io.Writer) error {
	_, err := w.Write([]byte(j.values.Encode()))
	return err
}

func (j *WWWFormEncode) Add(data interface{}) error {
	switch d := data.(type) {
	case map[string]string:
		for k, v := range d {
			j.values.Add(k, v)
		}
		return nil
	}
	return errors.New("Not Support www-form data type")
}

func (j *WWWFormEncode) Name() string {
	return "www-form"
}

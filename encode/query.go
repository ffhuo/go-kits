package encode

import (
	"errors"
	"io"
	"net/url"
)

type QueryEncode struct {
	values url.Values
}

func NewQueryEncode(data map[string]string) *QueryEncode {
	code := &QueryEncode{values: url.Values{}}
	for k, v := range data {
		code.values.Add(k, v)
	}
	return code
}

func (j *QueryEncode) Encode(w io.Writer) error {
	_, err := w.Write([]byte(j.values.Encode()))
	return err
}

func (j *QueryEncode) Add(data interface{}) error {
	switch d := data.(type) {
	case map[string]string:
		for k, v := range d {
			j.values.Add(k, v)
		}
		return nil
	}
	return errors.New("Not Support query data type")
}

func (j *QueryEncode) Name() string {
	return "query"
}

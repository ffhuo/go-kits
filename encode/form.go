package encode

import (
	"errors"
	"io"
	"mime/multipart"
)

type FormEncode struct {
	obj       map[string]string
	fileField string
	filename  string
}

func NewFormEncoderWithFiled(data map[string]string) *FormEncode {
	return &FormEncode{obj: data}
}

func NewFormEncoderWithFile(fieldname string, filename string) *FormEncode {
	return &FormEncode{
		obj:       map[string]string{},
		filename:  filename,
		fileField: fieldname,
	}
}

func (j *FormEncode) Encode(w io.Writer) error {
	p := multipart.NewWriter(w)
	for k, v := range j.obj {
		p.WriteField(k, v)
	}

	if j.filename != "" {
		_, err := p.CreateFormFile(j.fileField, j.filename)
		if err != nil {
			return err
		}
	}

	p.Close()
	return nil
}

func (j *FormEncode) Add(data interface{}) error {
	switch d := data.(type) {
	case map[string]string:
		for k, v := range d {
			j.obj[k] = v
		}
		return nil
	}
	return errors.New("Not Support Form data type")
}

func (j *FormEncode) Name() string {
	return "form"
}

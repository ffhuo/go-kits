package encode

import (
	"errors"
	"io"
)

type BodyEncode struct {
	reader io.Reader
}

func NewBodyEncoder(obj io.Reader) Encoder {
	if obj == nil {
		return nil
	}

	return &BodyEncode{reader: obj}
}

func (j *BodyEncode) Encode(w io.Writer) error {
	_, err := io.Copy(w, j.reader)
	return err
}

func (j *BodyEncode) Add(interface{}) error {
	return errors.New("Not Support Add Body data")
}

func (j *BodyEncode) Name() string {
	return "body"
}

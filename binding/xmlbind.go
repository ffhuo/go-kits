package binding

import (
	"encoding/xml"
	"fmt"
)

type xmlBinding struct{}

func (xmlBinding) Name() string {
	return "xml"
}

func (xmlBinding) Bind(body []byte, obj interface{}) error {
	if body == nil {
		return fmt.Errorf("invalid body")
	}
	if err := xml.Unmarshal(body, obj); err != nil {
		return err
	}
	return validate(obj)
}

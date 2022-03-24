package binding

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type tomlBinding struct{}

func (tomlBinding) Name() string {
	return "toml"
}

func (tomlBinding) Bind(body []byte, obj interface{}) error {
	if body == nil {
		return fmt.Errorf("invalid body")
	}
	if _, err := toml.Decode(string(body), obj); err != nil {
		return err
	}
	return validate(obj)
}

package binding

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type yamlBinding struct{}

func (yamlBinding) Name() string {
	return "yaml"
}

func (yamlBinding) Bind(body []byte, obj interface{}) error {
	if body == nil {
		return fmt.Errorf("invalid body")
	}
	if err := yaml.Unmarshal(body, obj); err != nil {
		return err
	}
	return validate(obj)
}

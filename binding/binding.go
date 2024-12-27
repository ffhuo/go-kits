// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import "strings"

// ContentType constants for different data formats
const (
	TYPEJSON = ".json"
	TYPEXML  = ".xml"
	TYPEYAML = ".yaml"
	TYPETOML = ".toml"
)

// Binding describes the interface which needs to be implemented for binding the
// data present in the request such as JSON request body, query parameters or
// the form POST.
type Binding interface {
	// Name returns the name of the binding implementation
	Name() string
	// Bind binds the passed []byte data to the passed interface{}
	Bind([]byte, interface{}) error
}

// StructValidator is the minimal interface which needs to be implemented for
// validation. It provides a default implementation using go-playground/validator.
type StructValidator interface {
	// ValidateStruct validates any struct type. It should never panic.
	// For slices/arrays, validation is performed on each element.
	// For non-struct types, validation is skipped and nil is returned.
	// For structs, full validation is performed and errors are returned if any.
	ValidateStruct(interface{}) error

	// Engine returns the underlying validator engine
	Engine() interface{}
}

// Validator is the default validator which implements the StructValidator
// interface using go-playground/validator
var Validator StructValidator = &defaultValidator{}

// Available binding implementations
var (
	JSON = jsonBinding{}
	XML  = xmlBinding{}
	YAML = yamlBinding{}
	TOML = tomlBinding{}
)

// Default returns the appropriate Binding instance based on the file extension
func Default(fileName string) Binding {
	switch strings.ToLower(fileName) {
	case TYPEJSON:
		return JSON
	case TYPEXML:
		return XML
	case TYPEYAML:
		return YAML
	case TYPETOML:
		return TOML
	default:
		return JSON
	}
}

// validate is a helper function to validate an interface using the default validator
func validate(obj interface{}) error {
	if Validator == nil {
		return nil
	}
	return Validator.ValidateStruct(obj)
}

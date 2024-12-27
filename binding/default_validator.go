// Copyright 2017 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

// defaultValidator implements the StructValidator interface using go-playground/validator
type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

// sliceValidateError represents validation errors for slice types
type sliceValidateError []error

// Error implements the error interface for sliceValidateError
func (err sliceValidateError) Error() string {
	var errMsgs []string
	for i, e := range err {
		if e == nil {
			continue
		}
		errMsgs = append(errMsgs, fmt.Sprintf("[%d]: %s", i, e.Error()))
	}
	return strings.Join(errMsgs, "\n")
}

// ValidateStruct implements the StructValidator interface
func (v *defaultValidator) ValidateStruct(obj interface{}) error {
	if obj == nil {
		return nil
	}

	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Ptr:
		return v.ValidateStruct(value.Elem().Interface())
	case reflect.Struct:
		return v.validateStruct(obj)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		validateRet := make(sliceValidateError, 0)
		for i := 0; i < count; i++ {
			if err := v.ValidateStruct(value.Index(i).Interface()); err != nil {
				validateRet = append(validateRet, err)
			}
		}
		if len(validateRet) > 0 {
			return validateRet
		}
		return nil
	default:
		return nil
	}
}

// validateStruct handles the actual validation of struct types
func (v *defaultValidator) validateStruct(obj interface{}) error {
	v.lazyinit()
	return v.validate.Struct(obj)
}

// Engine returns the underlying validator engine
func (v *defaultValidator) Engine() interface{} {
	v.lazyinit()
	return v.validate
}

// lazyinit initializes the validator instance if not already done
func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = validator.New()
	})
}

var _ StructValidator = &defaultValidator{}

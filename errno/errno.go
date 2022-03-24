package errno

import "fmt"

func Error(err error) *BusinessError {
	if berr, ok := err.(*BusinessError); ok {
		return berr
	}

	return &BusinessError{
		Code: -1,
		Msg:  err.Error(),
	}
}

type BusinessError struct {
	Code int
	Msg  string
	args []interface{}
}

func New(code int) *BusinessError {
	return &BusinessError{Code: code}
}

func Errorf(format string, a ...interface{}) *BusinessError {
	s := fmt.Sprintf(format, a...)
	return &BusinessError{
		Code: -1,
		Msg:  s,
	}
}

func NewWithArgs(code int, args ...interface{}) *BusinessError {
	return &BusinessError{
		Code: code,
		args: args,
	}
}

func (err BusinessError) ErrCode() int {
	return err.Code
}

func (err *BusinessError) SetErrMsg(format string, a ...interface{}) *BusinessError {
	err.Msg = fmt.Sprintf(format, a...)
	return err
}

func (err BusinessError) ErrMsg() string {
	return err.Msg
}

func (err BusinessError) ErrArgs() []interface{} {
	return err.args
}

func (err *BusinessError) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", err.Code, err.Msg)
}

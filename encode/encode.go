package encode

import "io"

// Encoder is the encoding interface
type Encoder interface {
	Encode(w io.Writer) error
	Add(interface{}) error
	Name() string
}

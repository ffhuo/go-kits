package gout

import (
	"context"
	"net/http"

	"github.com/ffhuo/go-kits/decode"
	"github.com/ffhuo/go-kits/encode"
)

// Req controls core data structure of http request
type Req struct {
	c   context.Context
	Err error

	method string
	url    string

	// http body
	bodyEncoder encode.Encoder
	bodyDecoder decode.Decoder

	// http header
	headerEncoder encode.Encoder
	headerDecoder decode.Decoder

	//cookie
	cookies []*http.Cookie

	req *http.Request

	cancel context.CancelFunc

	resp *http.Response

	res interface{}
}

package gout

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/ffhuo/go-kits/decode"
	"github.com/ffhuo/go-kits/encode"
)

var (
	defaultTimeout time.Duration = 10 * time.Second
)

// DataFlow controls core data structure of http request
type DataFlow struct {
	*http.Client

	c   context.Context
	Err error

	method   string
	url      string
	userName *string
	password *string

	// http body
	bodyEncoder encode.Encoder
	bodyDecoder decode.Decoder

	queryEncoder encode.Encoder

	// http header
	headerEncoder map[string]string

	//cookie
	cookies []*http.Cookie

	req *http.Request

	cancel context.CancelFunc

	resp *http.Response
}

func New(c ...*http.Client) *DataFlow {
	out := &DataFlow{}
	if len(c) == 0 || c[0] == nil {
		out.Client = http.DefaultClient
	} else {
		out.Client = c[0]
	}

	return out
}

func (d *DataFlow) Reset() {
	if d.cancel != nil {
		d.cancel()
	}
	d.Err = nil
	d.method = ""
	d.url = ""
	d.bodyDecoder = nil
	d.bodyEncoder = nil
	d.queryEncoder = nil

	d.headerEncoder = nil
	d.cookies = nil
	d.req = nil
	d.resp = nil
}

func (d *DataFlow) GET(url string) *DataFlow {
	d.method = http.MethodGet
	d.url = url
	return d
}

func (d *DataFlow) POST(url string) *DataFlow {
	d.method = http.MethodPost
	d.url = url
	return d
}

func (d *DataFlow) PUT(url string) *DataFlow {
	d.method = http.MethodPut
	d.url = url
	return d
}

func (d *DataFlow) DELETE(url string) *DataFlow {
	d.method = http.MethodDelete
	d.url = url
	return d
}

func (d *DataFlow) SetBasicAuth(userName, password string) *DataFlow {
	d.userName = &userName
	d.password = &password
	return d
}

func (d *DataFlow) SetHeader(header map[string]string) *DataFlow {
	d.headerEncoder = header
	return d
}

func (d *DataFlow) AddHeader(key, value string) *DataFlow {
	if len(d.headerEncoder) == 0 {
		d.headerEncoder = make(map[string]string)
	}
	d.headerEncoder[key] = value
	return d
}

func (d *DataFlow) SetRequest(req *http.Request) *DataFlow {
	d.req = req
	return d
}

func (d *DataFlow) SetCookies(cookies ...*http.Cookie) *DataFlow {
	d.cookies = append(d.cookies, cookies...)
	return d
}

func (d *DataFlow) AddCookies(cookie *http.Cookie) *DataFlow {
	d.cookies = append(d.cookies, cookie)
	return d
}

func (d *DataFlow) SetQuery(params map[string]string) *DataFlow {
	if d.queryEncoder == nil {
		d.queryEncoder = encode.NewQueryEncode(params)
	} else {
		d.Err = d.queryEncoder.Add(params)
	}
	return d
}

func (d *DataFlow) AddQuery(key, value string) *DataFlow {
	if d.queryEncoder == nil {
		d.queryEncoder = encode.NewQueryEncode(map[string]string{key: value})
	} else {
		d.Err = d.queryEncoder.Add(map[string]string{key: value})
	}
	return d
}

func (d *DataFlow) SetBody(body io.Reader) *DataFlow {
	d.bodyEncoder = encode.NewBodyEncoder(body)
	return d
}

func (d *DataFlow) SetFormWithFile(filedName, fileName string) *DataFlow {
	d.bodyEncoder = encode.NewFormEncoderWithFile(filedName, fileName)
	return d
}

func (d *DataFlow) SetForm(data map[string]string) *DataFlow {
	if d.bodyEncoder == nil {
		d.bodyEncoder = encode.NewFormEncoderWithFiled(data)
	} else {
		d.Err = d.bodyEncoder.Add(data)
	}
	return d
}

func (d *DataFlow) AddForm(key, value string) *DataFlow {
	if d.bodyEncoder == nil {
		d.bodyEncoder = encode.NewFormEncoderWithFiled(map[string]string{key: value})
	} else {
		d.Err = d.bodyEncoder.Add(map[string]string{key: value})
	}
	return d
}

func (d *DataFlow) SetJSON(data interface{}) *DataFlow {
	d.bodyEncoder = encode.NewJSONEncoder(data)
	return d
}

func (d *DataFlow) BindJSON(res interface{}) *DataFlow {
	d.bodyDecoder = decode.NewJSONDecode(res)
	return d
}

func (d *DataFlow) BindXML(res interface{}) *DataFlow {
	d.bodyDecoder = decode.NewXMLDecode(res)
	return d
}

func (d *DataFlow) BindYAML(res interface{}) *DataFlow {
	d.bodyDecoder = decode.NewYAMLDecode(res)
	return d
}

func (d *DataFlow) SetTimeout(timeout time.Duration) *DataFlow {
	d.Timeout = timeout
	return d
}

func (d *DataFlow) WithContext(ctx context.Context) *DataFlow {
	d.c = ctx
	return d
}

func (d *DataFlow) buildRequest() (*http.Request, error) {
	var (
		err error
		req *http.Request
	)
	body := &bytes.Buffer{}
	if d.req == nil {
		if d.bodyEncoder != nil {
			if err = d.bodyEncoder.Encode(body); d.Err != nil {
				return nil, err
			}
		}
		req, err = http.NewRequest(d.method, d.url, body)
		if err != nil {
			return nil, err
		}
	} else {
		req = d.req
		if len(d.method) > 0 {
			req.Method = d.method
		}
		if len(d.url) > 0 {
			req.URL, err = url.Parse(d.url)
			if err != nil {
				return nil, err
			}
		}
	}
	if d.queryEncoder != nil {
		if err = d.queryEncoder.Encode(body); err != nil {
			return nil, err
		}
		req.URL.RawQuery = body.String()
	}

	if d.c != nil {
		req = req.WithContext(d.c)
	}

	for _, cookie := range d.cookies {
		req.AddCookie(cookie)
	}

	if d.userName != nil && d.password != nil {
		req.SetBasicAuth(*d.userName, *d.password)
	}

	return req, nil
}

func (d *DataFlow) buildHeader() (http.Header, error) {
	header := http.Header{}
	for k, v := range d.headerEncoder {
		header.Set(k, v)
	}
	if d.bodyEncoder != nil {
		switch d.bodyEncoder.Name() {
		case "json":
			header.Set("Content-Type", "application/json")
		case "form":
			header.Set("Content-Type", "multipart/form-data")
		case "www-form":
			header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}
	return header, nil
}

func (d *DataFlow) decodeBody() ([]byte, error) {
	var (
		err       error
		bodyBytes []byte
	)
	bodyBytes, err = ioutil.ReadAll(d.resp.Body)
	if err != nil {
		return nil, err
	}
	d.resp.Body.Close()
	if d.bodyDecoder == nil {
		return bodyBytes, nil
	}
	if len(bodyBytes) > 0 {
		d.resp.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	}

	if err = d.bodyDecoder.Decode(d.resp.Body); err != nil {
		return bodyBytes, err
	}
	return bodyBytes, nil
}

func (d *DataFlow) Do() ([]byte, error) {
	if d.Err != nil {
		return nil, d.Err
	}

	defer d.Reset()

	d.req, d.Err = d.buildRequest()
	if d.Err != nil {
		return nil, d.Err
	}

	d.req.Header, d.Err = d.buildHeader()
	if d.Err != nil {
		return nil, d.Err
	}

	if d.Timeout == 0 {
		d.Timeout = defaultTimeout
	}

	var ctx context.Context
	if d.c != nil {
		ctx = d.c
	} else {
		ctx = context.Background()
	}
	d.c, d.cancel = context.WithTimeout(ctx, d.Timeout)
	defer d.cancel()

	d.resp, d.Err = d.Client.Do(d.req)
	if d.Err != nil {
		return nil, d.Err
	}

	return d.decodeBody()
}

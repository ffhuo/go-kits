package gout

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type Gout struct {
	*http.Client
	Req
}

func New(c ...*http.Client) *Gout {
	out := &Gout{}
	if len(c) == 0 || c[0] == nil {
		out.Client = http.DefaultClient
	} else {
		out.Client = c[0]
	}

	return out
}

// GET send HTTP GET method
func GET(url string) *Gout {
	return New().GET(url)
}

// POST send HTTP POST method
func POST(url string) *Gout {
	return New().POST(url)
}

// PUT send HTTP PUT method
func PUT(url string) *Gout {
	return New().PUT(url)
}

// DELETE send HTTP DELETE method
func DELETE(url string) *Gout {
	return New().DELETE(url)
}

func (g *Gout) GET(url string) *Gout {
	g.method = "GET"
	g.url = url
	return g
}

func (g *Gout) POST(url string) *Gout {
	g.method = "POST"
	g.url = url
	return g
}

func (g *Gout) PUT(url string) *Gout {
	g.method = "PUT"
	g.url = url
	return g
}

func (g *Gout) DELETE(url string) *Gout {
	g.method = "DELETE"
	g.url = url
	return g
}

func (g *Gout) SetHeader(header map[string]string) *Gout {
	g.header = header
	return g
}

func (g *Gout) AddHeader(key, value string) *Gout {
	if len(g.header) == 0 {
		g.header = make(map[string]string)
	}
	g.header[key] = value
	return g
}

func (g *Gout) SetTimeout(timeout time.Duration) *Gout {
	g.Timeout = timeout
	return g
}

func (g *Gout) SetRequest(req *http.Request) *Gout {
	g.req = req
	return g
}

func (g *Gout) SetCookies(cookies []*http.Cookie) *Gout {
	g.cookies = cookies
	return g
}

func (g *Gout) AddCookies(cookie *http.Cookie) *Gout {
	g.cookies = append(g.cookies, cookie)
	return g
}

func (g *Gout) SetQuery(params map[string]string) *Gout {
	if len(g.query) == 0 {
		g.query = make(map[string]string)
	}
	g.query = params
	return g
}

func (g *Gout) AddQuery(key, value string) *Gout {
	if len(g.query) == 0 {
		g.query = make(map[string]string)
	}
	g.query[key] = value
	return g
}

func (g *Gout) SetBody(body io.Reader) *Gout {
	g.body = body
	return g
}

func (g *Gout) SetForm(form io.Reader) *Gout {
	g.form = form
	return g
}

func (g *Gout) SetJSON(data interface{}) *Gout {
	body, _ := json.Marshal(data)
	g.body = bytes.NewReader(body)
	return g
}

func (g *Gout) BindJSON(res interface{}) *Gout {
	g.res = res
	return g
}

func (g *Gout) Do() ([]byte, error) {
	var body io.Reader
	if g.form != nil {
		body = g.form
	} else {
		body = g.body
	}

	if g.req == nil {
		g.req, g.Err = http.NewRequest(g.method, g.url, body)
		if g.Err != nil {
			return nil, g.Err
		}
	}

	for k, v := range g.header {
		g.req.Header.Set(k, v)
	}

	if len(g.cookies) > 0 {
		for _, cookie := range g.cookies {
			g.req.AddCookie(cookie)
		}
	}

	if len(g.query) > 0 {
		q := g.req.URL.Query()
		for k, v := range g.query {
			q.Add(k, v)
		}
		g.req.URL.RawQuery = q.Encode()
	}

	if g.form != nil {
		g.req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if g.body != nil {
		g.req.Header.Set("Content-Type", "application/json")
	}

	if g.Timeout == 0 {
		g.Timeout = time.Second * 10
	}

	g.c, g.cancel = context.WithTimeout(context.Background(), g.Timeout)
	defer g.cancel()

	g.resp, g.Err = g.Client.Do(g.req.WithContext(g.c))
	if g.Err != nil {
		return nil, g.Err
	}

	var bodyBytes []byte
	bodyBytes, g.Err = ioutil.ReadAll(g.resp.Body)
	if g.Err != nil {
		return nil, g.Err
	}

	g.resp.Body.Close()
	if g.res != nil {
		g.Err = json.Unmarshal(bodyBytes, g.res)
	}
	return bodyBytes, g.Err
}

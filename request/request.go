package request

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

func SendRequest(req *http.Request) (res []byte, err error) {
	// 3秒请求超时
	client := &http.Client{Timeout: time.Duration(10 * time.Second)}
	resp, err := client.Do(req)
	if err != nil {
		return res, errors.Wrapf(err, "request to %s failed", req.URL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return res, fmt.Errorf("request to %s failed: %d", req.URL, resp.StatusCode)
	}

	res, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	return
}

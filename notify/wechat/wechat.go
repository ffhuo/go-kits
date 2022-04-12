package wechat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ffhuo/go-kits/request"
	"github.com/pkg/errors"
)

type Client struct {
	corpid     string
	corpsecret string
	token      string
	expireTime time.Time
}

func New(corpid string, corpsecret string) *Client {
	return &Client{
		corpid:     corpid,
		corpsecret: corpsecret,
	}
}

func (c *Client) GetAccessToken() error {
	if c.token != "" && c.expireTime.After(time.Now()) {
		return nil
	}

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", c.corpid, c.corpsecret)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := request.SendRequest(req)
	if err != nil {
		return err
	}

	var result struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}

	if err = json.Unmarshal(resp, &result); err != nil {
		return errors.Wrap(err, "failed to unmarshal response")
	}
	if result.ErrCode != 0 {
		return errors.Errorf("failed to get user by mobile: %s", result.ErrMsg)
	}
	c.token = result.AccessToken
	c.expireTime = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	return nil
}

package wechat

import (
	"fmt"
	"time"

	"github.com/ffhuo/go-kits/gout"
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

	var result struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}

	if _, err := gout.GET("https://qyapi.weixin.qq.com/cgi-bin/gettoken").
		AddQuery("corpid", c.corpid).
		AddQuery("corpsecret", c.corpsecret).
		BindJSON(&result).
		Do(); err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("failed to get user by mobile: %s", result.ErrMsg)
	}
	c.token = result.AccessToken
	c.expireTime = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	return nil
}

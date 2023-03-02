package feishu

import (
	"encoding/json"
	"time"

	"github.com/ffhuo/go-kits/gout"
)

// Client 端
type Client struct {
	appId     string
	appSecret string
	Token     *AppAccessToken
}

// NewClient ...
func New(appId, appSecret string) *Client {
	client := &Client{
		appId:     appId,
		appSecret: appSecret,
	}
	return client
}

// GetUserInfo 根据邮件或手机号获取用户信息
func (client *Client) GetUserInfo(emails []string, phones []string) (*AppUserInfo, error) {
	var (
		err   error
		query string
		v     AppUserInfo
	)

	for _, email := range emails {
		query += "&emails=" + email
	}

	for _, phone := range phones {
		query += "&mobiles=" + phone
	}

	query = query[1:]
	url := ApiGetUserByEmail + "?" + query

	_, err = gout.GET(url).
		AddHeader("Authorization", "Bearer "+client.Token.TenantAccessToken).
		BindJSON(&v).Do()
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// SendWebhookMessage 发送webhook文本信息
func (client *Client) SendWebhookMessage(url string, subject, content string, atAll bool) error {
	if err := client.getAppAccessToken(); err != nil {
		return err
	}

	bodyContent := content
	if atAll {
		bodyContent = `<at user_id="all">所有人</at> ` + content
	}
	reqBody := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]interface{}{
			"text": bodyContent,
		},
	}
	_, err := gout.POST(url).
		Debug().
		SetJSON(reqBody).Do()
	if err != nil {
		return err
	}

	return nil
}

// SendMessage 发送富文本消息
func (client *Client) SendMessage(openID string, subject, content string) error {
	if err := client.getAppAccessToken(); err != nil {
		return err
	}
	var con [][]AppContentData
	if err := json.Unmarshal([]byte(content), &con); err != nil {
		return err
	}

	reqBody := map[string]interface{}{
		"open_id":  openID,
		"msg_type": "post",
		"content": map[string]interface{}{
			"post": map[string]interface{}{
				"zh_cn": map[string]interface{}{
					"title":   subject,
					"content": con,
				},
			},
		},
	}

	_, err := gout.POST(ApiRobotSendMessage).
		Debug().
		AddHeader("Authorization", "Bearer "+client.Token.TenantAccessToken).
		SetJSON(reqBody).Do()
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) getAppAccessToken() error {
	if client.Token != nil && client.Token.Expire < time.Now().Unix() {
		return nil
	}
	reqBody := map[string]interface{}{
		"app_id":     client.appId,
		"app_secret": client.appSecret,
	}

	res := AppAccessToken{}
	_, err := gout.POST(ApiAppAccessTokenInternal).
		SetJSON(reqBody).
		BindJSON(&res).Do()
	if err != nil {
		return err
	}
	client.Token = &res
	return nil
}

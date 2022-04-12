package dingding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	dingtalk "github.com/alibabacloud-go/dingtalk/oauth2_1_0"
	dingrobot "github.com/alibabacloud-go/dingtalk/robot_1_0"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/pkg/errors"
)

type Client struct {
	appKey     string
	appSecret  string
	token      *string
	expireTime time.Time
}

func New(appKey string, appSecret string) *Client {
	return &Client{
		appKey:    appKey,
		appSecret: appSecret,
	}
}

func (c *Client) initAuthClient() (*dingtalk.Client, error) {
	config := &openapi.Config{}
	config.Protocol = tea.String("https")
	config.RegionId = tea.String("central")

	return dingtalk.NewClient(config)
}

func (c *Client) GetAccessToken() error {
	cli, err := c.initAuthClient()
	if err != nil {
		return err
	}

	if c.token != nil && c.expireTime.After(time.Now()) {
		return nil
	}

	req := &dingtalk.GetAccessTokenRequest{
		AppKey:    tea.String(c.appKey),
		AppSecret: tea.String(c.appSecret),
	}

	resp, err := cli.GetAccessToken(req)
	if err != nil {
		return err
	}

	c.token = resp.Body.AccessToken
	c.expireTime = time.Now().Add(time.Duration(*resp.Body.ExpireIn) * time.Second)
	return nil
}

func (c *Client) initRobotClient() (*dingrobot.Client, error) {
	config := &openapi.Config{}
	config.Protocol = tea.String("https")
	config.RegionId = tea.String("central")

	return dingrobot.NewClient(config)
}

// SendRobotMessage 发送消息
// 仅支持markdown
// {"text": "hello text","title": "hello title"}
func (c *Client) SendRobotMessage(userIds []*string, msg interface{}) error {
	cli, err := c.initRobotClient()
	if err != nil {
		return err
	}
	if err = c.GetAccessToken(); err != nil {
		return err
	}

	header := &dingrobot.BatchSendOTOHeaders{}
	header.XAcsDingtalkAccessToken = tea.String(*c.token)

	body, _ := json.Marshal(msg)
	req := &dingrobot.BatchSendOTORequest{
		RobotCode: tea.String(c.appKey),
		UserIds:   userIds,
		MsgKey:    tea.String("sampleMarkdown"),
		MsgParam:  tea.String(string(body)),
	}
	_, err = cli.BatchSendOTOWithOptions(req, header, &util.RuntimeOptions{})
	return err
}

func (c *Client) GetUserByMobile(mobile string) (string, error) {
	var err error
	if err = c.GetAccessToken(); err != nil {
		return "", err
	}
	url := "https://oapi.dingtalk.com/topapi/v2/user/getbymobile?access_token=" + *c.token

	args := make(map[string]interface{}, 2)
	args["mobile"] = mobile
	args["support_exclusive_account_search"] = "true"
	b, _ := json.Marshal(args)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}

	res, err := c.sendRequest(req)
	if err != nil {
		return "", err
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		Result  struct {
			UserId string `json:"userid"`
		}
	}

	if err = json.Unmarshal(res, &result); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal response")
	}
	if result.ErrCode != 0 {
		return "", errors.Errorf("failed to get user by mobile: %s", result.ErrMsg)
	}
	return result.Result.UserId, nil
}

func (c *Client) sendRequest(req *http.Request) (res []byte, err error) {
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

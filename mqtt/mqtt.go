package mqtt

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Option func(*option)

type option struct {
	brokers           []string
	clientId          string
	username          string
	password          string
	publishHandler    mqtt.MessageHandler
	connectHandler    mqtt.OnConnectHandler
	disconnectHandler mqtt.ConnectionLostHandler
}

func WithBroker(brokers []string) Option {
	return func(opt *option) {
		opt.brokers = brokers
	}
}

func WithClientId(clientId string) Option {
	return func(opt *option) {
		opt.clientId = clientId
	}
}

func WithAuth(username, password string) Option {
	return func(opt *option) {
		opt.username = username
		opt.password = password
	}
}

func WithPublishHandler(handler mqtt.MessageHandler) Option {
	return func(opt *option) {
		opt.publishHandler = handler
	}
}

func WithConnectHandler(handler mqtt.OnConnectHandler) Option {
	return func(opt *option) {
		opt.connectHandler = handler
	}
}

func WithDiconnectHandler(handler mqtt.ConnectionLostHandler) Option {
	return func(opt *option) {
		opt.disconnectHandler = handler
	}
}

type Client struct {
	cli mqtt.Client
}

func New(opts ...Option) (*Client, error) {
	options := NewClientOptions(opts...)
	cli := mqtt.NewClient(options)
	if token := cli.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return &Client{cli: cli}, nil
}

func (c *Client) Close() {
	if c.cli == nil {
		return
	}
	c.cli.Disconnect(250)
}

func NewClientOptions(opts ...Option) *mqtt.ClientOptions {
	var opt option
	for _, o := range opts {
		o(&opt)
	}

	options := mqtt.NewClientOptions()
	for _, broker := range opt.brokers {
		options.AddBroker(broker)
	}
	if opt.clientId != "" {
		options.SetClientID(opt.clientId)
	} else {
		options.SetClientID("mqtt-client")
	}
	options.SetUsername(opt.username)
	options.SetPassword(opt.password)
	options.SetDefaultPublishHandler(opt.publishHandler)
	options.SetOnConnectHandler(opt.connectHandler)
	options.SetConnectionLostHandler(opt.disconnectHandler)
	return options
}

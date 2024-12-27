package etcd

/*
import (
	"context"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// Option custom setup config
type Option func(*option)

type option struct {
	endpoints []string
	timeout   time.Duration
	username  string
	password  string
}

func WithTimeount(timeout time.Duration) Option {
	return func(opt *option) {
		opt.timeout = timeout
	}
}

func WithEndpoint(endpoints []string) Option {
	return func(opt *option) {
		opt.endpoints = endpoints
	}
}

func WithUser(username, password string) Option {
	return func(opt *option) {
		opt.username = username
		opt.password = password
	}
}

type Client struct {
	cli     *clientv3.Client
	timeout time.Duration
}

func New(opts ...Option) (*Client, error) {
	var opt option
	for _, o := range opts {
		o(&opt)
	}
	config := clientv3.Config{
		Endpoints:   opt.endpoints,
		DialTimeout: opt.timeout,
	}
	if opt.username != "" && opt.password != "" {
		config.Username = opt.username
		config.Password = opt.password
	}
	cli, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}
	return &Client{cli: cli, timeout: opt.timeout}, nil
}

func (c *Client) Put(key, value string, opts ...clientv3.OpOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	_, err := c.cli.Put(ctx, key, value, opts...)
	return err
}

func (c *Client) Get(key string, opts ...clientv3.OpOption) ([]*mvccpb.KeyValue, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	var result []*mvccpb.KeyValue
	for {
		resp, err := c.cli.Get(ctx, key, opts...)
		if err != nil {
			return nil, err
		}
		if len(resp.Kvs) > 0 {
			result = append(result, resp.Kvs...)
		}
		if !resp.More {
			break
		}
	}

	return result, nil
}

func (c *Client) Delete(key string, opts ...clientv3.OpOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	_, err := c.cli.Delete(ctx, key, opts...)
	return err
}

func (c *Client) Close(ctx context.Context) error {
	if c.cli == nil {
		return nil
	}
	return c.cli.Close()
}

func (c *Client) Lock(lockKey string) (*concurrency.Mutex, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	session, err := concurrency.NewSession(c.cli)
	if err != nil {
		return nil, err
	}
	mu := concurrency.NewMutex(session, lockKey)
	if err = mu.Lock(ctx); err != nil {
		return mu, err
	}
	return mu, nil
}

func (c *Client) Unlock(mu *concurrency.Mutex) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	return mu.Unlock(ctx)
}
*/

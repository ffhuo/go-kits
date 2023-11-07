package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCli struct {
	ctx           context.Context
	isCluster     bool
	clusterClient *redis.ClusterClient
	signalClient  *redis.Client
}

func New(ctx context.Context, addrs []string, password string) (*RedisCli, error) {
	var err error
	c := &RedisCli{
		ctx: ctx,
	}

	if len(addrs) == 1 {
		c.signalClient, err = c.redisConnect(addrs, password)
	} else {
		c.clusterClient, err = c.redisClusterConnect(addrs, password)
		c.isCluster = true
	}
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *RedisCli) redisConnect(addrs []string, password string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addrs[0],
		Password:     password,
		MaxRetries:   3,
		PoolSize:     10,
		MinIdleConns: 10,
	})

	if err := client.Ping(c.ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis err: %v", err)
	}

	return client, nil
}

func (c *RedisCli) redisClusterConnect(addrs []string, password string) (*redis.ClusterClient, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        addrs,
		Password:     password,
		MaxRetries:   3,
		PoolSize:     10,
		MinIdleConns: 10,
	})

	if err := client.Ping(c.ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis err: %v", err)
	}

	return client, nil
}

func (c *RedisCli) RedisClient() redis.Cmdable {
	if c.isCluster {
		return c.clusterClient
	}
	return c.signalClient
}

// Set set some <key,value> into redis
func (c *RedisCli) Set(key, value string, ttl time.Duration) error {
	if err := c.RedisClient().Set(c.ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set key: %s err: %v", key, err)
	}

	return nil
}

func (c *RedisCli) SetWithData(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("redis set key: %s marshal err: %v", key, err)
	}
	return c.Set(key, string(data), ttl)
}

// Get get some key from redis
func (c *RedisCli) Get(key string) (string, error) {
	value, err := c.RedisClient().Get(c.ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("redis get key: %s err: %v", key, err)
	}

	return value, nil
}

// SetNX
func (c *RedisCli) SetNX(key, value string, ttl time.Duration) bool {
	ok, _ := c.RedisClient().SetNX(c.ctx, key, value, ttl).Result()
	return ok
}

// TTL get some key from redis
func (c *RedisCli) TTL(key string) (time.Duration, error) {
	ttl, err := c.RedisClient().TTL(c.ctx, key).Result()
	if err != nil {
		return -1, fmt.Errorf("redis get key: %s err: %v", key, err)
	}

	return ttl, nil
}

// Expire expire some key
func (c *RedisCli) Expire(key string, ttl time.Duration) bool {
	ok, _ := c.RedisClient().Expire(c.ctx, key, ttl).Result()
	return ok
}

// ExpireAt expire some key at some time
func (c *RedisCli) ExpireAt(key string, ttl time.Time) bool {
	ok, _ := c.RedisClient().ExpireAt(c.ctx, key, ttl).Result()
	return ok
}

func (c *RedisCli) Exists(keys ...string) bool {
	if len(keys) == 0 {
		return true
	}
	value, _ := c.RedisClient().Exists(c.ctx, keys...).Result()
	return value > 0
}

func (c *RedisCli) Del(key string) bool {
	if key == "" {
		return true
	}

	value, _ := c.RedisClient().Del(c.ctx, key).Result()
	return value > 0
}

func (c *RedisCli) Incr(key string) int64 {
	value, _ := c.RedisClient().Incr(c.ctx, key).Result()
	return value
}

func (c *RedisCli) Subscribe(channel string) *redis.PubSub {
	if c.isCluster {
		return c.clusterClient.Subscribe(c.ctx, channel)
	}
	return c.signalClient.Subscribe(c.ctx, channel)
}

func (c *RedisCli) Publish(channel string, data string) error {
	var err error
	if c.isCluster {
		_, err = c.clusterClient.Publish(c.ctx, channel, data).Result()
		return err
	}
	_, err = c.signalClient.Publish(c.ctx, channel, data).Result()
	return err
}

func (c *RedisCli) PublishWithData(channel string, data interface{}) error {
	str, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.Publish(channel, string(str))
}

// Sadd set 添加元素
func (c *RedisCli) Sadd(key string, members ...interface{}) error {
	_, err := c.RedisClient().SAdd(c.ctx, key, members...).Result()
	return err
}

func (c *RedisCli) Srem(key string, members ...interface{}) int64 {
	value, _ := c.RedisClient().SRem(c.ctx, key, members...).Result()
	return value
}

// SIsMember set 判断元素是否存在
func (c *RedisCli) SIsMember(key string, member interface{}) bool {
	value, _ := c.RedisClient().SIsMember(c.ctx, key, member).Result()
	return value
}

// Scard set 获取元素数量
func (c *RedisCli) Scard(key string) int64 {
	value, _ := c.RedisClient().SCard(c.ctx, key).Result()
	return value
}

// Hset hash 设置key中values[0]值values[1],values[2]值values[3]
// redis 版本v4
func (c *RedisCli) Hset(key string, values ...interface{}) error {
	_, err := c.RedisClient().HSet(c.ctx, key, values...).Result()
	return err
}

// HMSet hash 设置key中values[0]值values[1],values[2]值values[3]
// redis 版本v3
func (c *RedisCli) Hmset(key string, values ...interface{}) error {
	_, err := c.RedisClient().HMSet(c.ctx, key, values...).Result()
	return err
}

// Hget hash get key filed value
func (c *RedisCli) Hget(key, field string) (string, error) {
	return c.RedisClient().HGet(c.ctx, key, field).Result()
}

// HgetAll hash get key all field and value
func (c *RedisCli) HgetAll(key string) (map[string]string, error) {
	return c.RedisClient().HGetAll(c.ctx, key).Result()
}

func (c *RedisCli) Hdel(key string, field ...string) int64 {
	value, _ := c.RedisClient().HDel(c.ctx, key, field...).Result()
	return value
}

// Close close redis client
func (c *RedisCli) Close() error {
	if c.clusterClient != nil {
		c.clusterClient.Close()
	}
	if c.signalClient != nil {
		c.signalClient.Close()
	}
	return nil
}

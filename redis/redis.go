package redis

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/pkg/errors"
)

type RedisCli struct {
	isCluster     bool
	clusterClient *redis.ClusterClient
	signalClient  *redis.Client
}

func New(addrs []string, password string) (*RedisCli, error) {
	var err error
	c := &RedisCli{}

	if len(addrs) == 1 {
		c.signalClient, err = redisConnect(addrs, password)
	} else {
		c.clusterClient, err = redisClusterConnect(addrs, password)
		c.isCluster = true
	}
	if err != nil {
		return nil, err
	}

	return c, nil
}

func redisConnect(addrs []string, password string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addrs[0],
		Password:     password,
		MaxRetries:   3,
		PoolSize:     10,
		MinIdleConns: 10,
	})

	if err := client.Ping().Err(); err != nil {
		return nil, errors.Wrap(err, "ping redis err")
	}

	return client, nil
}

func redisClusterConnect(addrs []string, password string) (*redis.ClusterClient, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        addrs,
		Password:     password,
		MaxRetries:   3,
		PoolSize:     10,
		MinIdleConns: 10,
	})

	if err := client.Ping().Err(); err != nil {
		return nil, errors.Wrap(err, "ping redis err")
	}

	return client, nil
}

func (c *RedisCli) getRedisClient() redis.Cmdable {
	if c.isCluster {
		return c.clusterClient
	}
	return c.signalClient
}

// Set set some <key,value> into redis
func (c *RedisCli) Set(key, value string, ttl time.Duration) error {
	if err := c.getRedisClient().Set(key, value, ttl).Err(); err != nil {
		return errors.Wrapf(err, "redis set key: %s err", key)
	}

	return nil
}

func (c *RedisCli) SetWithData(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return errors.Wrapf(err, "redis set key: %s marshal err", key)
	}
	return c.Set(key, string(data), ttl)
}

// Get get some key from redis
func (c *RedisCli) Get(key string) (string, error) {
	value, err := c.getRedisClient().Get(key).Result()
	if err != nil {
		return "", errors.Wrapf(err, "redis get key: %s err", key)
	}

	return value, nil
}

// SetNX
func (c *RedisCli) SetNX(key, value string, ttl time.Duration) bool {
	ok, _ := c.getRedisClient().SetNX(key, value, ttl).Result()
	return ok
}

// TTL get some key from redis
func (c *RedisCli) TTL(key string) (time.Duration, error) {
	ttl, err := c.getRedisClient().TTL(key).Result()
	if err != nil {
		return -1, errors.Wrapf(err, "redis get key: %s err", key)
	}

	return ttl, nil
}

// Expire expire some key
func (c *RedisCli) Expire(key string, ttl time.Duration) bool {
	ok, _ := c.getRedisClient().Expire(key, ttl).Result()
	return ok
}

// ExpireAt expire some key at some time
func (c *RedisCli) ExpireAt(key string, ttl time.Time) bool {
	ok, _ := c.getRedisClient().ExpireAt(key, ttl).Result()
	return ok
}

//
func (c *RedisCli) Exists(keys ...string) bool {
	if len(keys) == 0 {
		return true
	}
	value, _ := c.getRedisClient().Exists(keys...).Result()
	return value > 0
}

func (c *RedisCli) Del(key string) bool {
	if key == "" {
		return true
	}

	value, _ := c.getRedisClient().Del(key).Result()
	return value > 0
}

func (c *RedisCli) Incr(key string) int64 {
	value, _ := c.getRedisClient().Incr(key).Result()
	return value
}

func (c *RedisCli) Subscribe(channel string) *redis.PubSub {
	if c.isCluster {
		return c.clusterClient.Subscribe(channel)
	}
	return c.signalClient.Subscribe(channel)
}

func (c *RedisCli) Publish(channel string, data string) error {
	var err error
	if c.isCluster {
		_, err = c.clusterClient.Publish(channel, data).Result()
		return err
	}
	_, err = c.signalClient.Publish(channel, data).Result()
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
	_, err := c.getRedisClient().SAdd(key, members...).Result()
	return err
}

func (c *RedisCli) Srem(key string, members ...interface{}) int64 {
	value, _ := c.getRedisClient().SRem(key, members...).Result()
	return value
}

// SIsMember set 判断元素是否存在
func (c *RedisCli) SIsMember(key string, member interface{}) bool {
	value, _ := c.getRedisClient().SIsMember(key, member).Result()
	return value
}

// Scard set 获取元素数量
func (c *RedisCli) Scard(key string) int64 {
	value, _ := c.getRedisClient().SCard(key).Result()
	return value
}

// Hset hash 设置key中values[0]值values[1],values[2]值values[3]
// redis 版本v4
func (c *RedisCli) Hset(key string, values ...interface{}) error {
	_, err := c.getRedisClient().HSet(key, values...).Result()
	return err
}

// HMSet hash 设置key中values[0]值values[1],values[2]值values[3]
// redis 版本v3
func (c *RedisCli) Hmset(key string, values ...interface{}) error {
	_, err := c.getRedisClient().HMSet(key, values...).Result()
	return err
}

// Hget hash get key filed value
func (c *RedisCli) Hget(key, field string) (string, error) {
	return c.getRedisClient().HGet(key, field).Result()
}

// HgetAll hash get key all field and value
func (c *RedisCli) HgetAll(key string) (map[string]string, error) {
	return c.getRedisClient().HGetAll(key).Result()
}

func (c *RedisCli) Hdel(key string, field ...string) int64 {
	value, _ := c.getRedisClient().HDel(key, field...).Result()
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

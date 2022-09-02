package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	redisv8 "github.com/go-redis/redis/v8"
)

type Client interface {
	Set(key string, obj interface{}, ttl time.Duration) error
	Get(key string, obj interface{}) error
	Delete(key string) error
	Keys(pattern string) ([]string, error)
}

type redis struct {
	client     *redisv8.Client
	expiration time.Duration
}

type Options struct {
	Address           string
	Password          string
	Db                int
	MaxRetries        int
	DefaultExpiration time.Duration
}

func NewClient(opt Options) (Client, error) {
	opts := &redisv8.Options{
		Addr:       opt.Address,
		Password:   opt.Password,
		DB:         opt.Db,
		MaxRetries: opt.MaxRetries,
	}

	client := redisv8.NewClient(opts)
	if err := ping(client); err != nil {
		return nil, err
	}

	return &redis{
		client:     client,
		expiration: opt.DefaultExpiration,
	}, nil
}

func (r *redis) Set(key string, obj interface{}, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = -1
	}

	if obj == nil || key == "" {
		return nil
	}

	val, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = r.client.Set(context.Background(), key, val, ttl).Result()
	return err
}

func (r *redis) Get(key string, obj interface{}) error {
	val, err := r.client.Get(context.Background(), key).Result()
	if errors.Is(err, redisv8.Nil) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), obj)
}

func (r *redis) Keys(pattern string) ([]string, error) {
	keys, err := r.client.Keys(context.Background(), pattern).Result()
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (r *redis) Delete(key string) error {
	_, err := r.client.Del(context.Background(), key).Result()
	return err
}

func ping(c *redisv8.Client) error {
	_, err := c.Ping(context.Background()).Result()
	return err
}
